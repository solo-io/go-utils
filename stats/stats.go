package stats

import (
	"context"
	"fmt"
	"os"
	"sort"
	"text/template"
	"time"

	"net/http"
	"net/http/pprof"

	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/solo-io/go-utils/contextutils"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/zpages"
	"go.uber.org/zap"
)

const (
	DefaultEnvVar       = "START_STATS_SERVER"
	DefaultEnabledValue = "true"
	DefaultPort         = 9091
)

type StartupOptions struct {
	// only start the server if this env var is present in the environment, and it is set to the given value
	// a StartStatsServer invocation when this is not the case is a no-op
	// if EnvVar is not provided, then the server starts unconditionally
	EnvVar       string
	EnabledValue string

	// listen on this port
	Port int

	// If set, the server will use this `AtomicLevel` to serve
	// the "/logging" endpoint instead of building its own logger.
	LogLevel *zap.AtomicLevel
}

// return options indicating that the server should:
//   - start up only if DefaultEnvVar is set to DefaultEnabledValue
//   - listen on DefaultPort
func DefaultStartupOptions() StartupOptions {
	return StartupOptions{
		EnvVar:       DefaultEnvVar,
		EnabledValue: DefaultEnabledValue,
		Port:         DefaultPort,
	}
}

// start the server with the default startup options
func ConditionallyStartStatsServer(addhandlers ...func(mux *http.ServeMux, profiles map[string]string)) {
	StartStatsServerWithPort(DefaultStartupOptions(), addhandlers...)
}

func StartStatsServerWithPort(startupOpts StartupOptions, addhandlers ...func(mux *http.ServeMux, profiles map[string]string)) {
	StartCancellableStatsServerWithPort(context.Background(), startupOpts, addhandlers...)
}

func StartCancellableStatsServerWithPort(ctx context.Context, startupOpts StartupOptions, customAddHandlers ...func(mux *http.ServeMux, profiles map[string]string)) {
	// if the env var was provided (i.e., startup is conditional) and the value of that env var is not the expected value, then return and do nothing
	if startupOpts.EnvVar != "" && os.Getenv(startupOpts.EnvVar) != startupOpts.EnabledValue {
		return
	}

	if envLogLevel := os.Getenv(contextutils.LogLevelEnvName); envLogLevel != "" {
		contextutils.SetLogLevelFromString(envLogLevel)
	}

	go RunCancellableGoroutineStat(ctx)

	// The running instance of the Stats server
	var server *http.Server

	addHandlers := append(customAddHandlers, addPprof, addStats)

	// Run the server in a goroutine
	go func() {
		mux := new(http.ServeMux)

		mux.Handle("/logging", getLoggingHandler(startupOpts))

		for _, addHandler := range addHandlers {
			addHandler(mux, profileDescriptions)
		}

		// add the index
		mux.HandleFunc("/", Index)

		server = &http.Server{
			Addr:    fmt.Sprintf(":%d", startupOpts.Port),
			Handler: mux,
		}
		contextutils.LoggerFrom(ctx).Infof("Stats server starting at %s", server.Addr)
		err := server.ListenAndServe()
		if err == http.ErrServerClosed {
			contextutils.LoggerFrom(ctx).Infof("Stats server closed")
		} else {
			contextutils.LoggerFrom(ctx).Warnf("Stats server closed with unexpected error: %v", err)
		}
	}()

	// Run a separate goroutine to handle the server shutdown when the context is cancelled
	go func() {
		<-ctx.Done()
		if server != nil {
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer shutdownCancel()
			if err := server.Shutdown(shutdownCtx); err != nil {
				contextutils.LoggerFrom(shutdownCtx).Warnf("Stats server shutdown returned error: %v", err)
			}
		}
	}()
}

func getLoggingHandler(startupOpts StartupOptions) zap.AtomicLevel {
	// If the AtomicLevel is configured in StartupOptions, respect that
	if startupOpts.LogLevel != nil {
		return *startupOpts.LogLevel
	}

	// Fallback to the existing AtomicLevel
	return contextutils.GetLogHandler()
}

func addPprof(mux *http.ServeMux, profiles map[string]string) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	profiles["/debug/pprof/"] = `PProf related things:<br/>
	<a href="/debug/pprof/goroutine?debug=2">full goroutine stack dump</a>
	`
}

func addStats(mux *http.ServeMux, profiles map[string]string) {
	exporter, err := prometheus.NewExporter(prometheus.Options{})
	if err == nil {
		view.RegisterExporter(exporter)
		mux.Handle("/metrics", exporter)

		profiles["/metrics"] = "Prometheus format metrics"
	}

	zpages.Handle(mux, "/zpages")
	profiles["/zpages"] = `Tracing. See <a href="/zpages/tracez">list of spans</a>`
}

func Index(w http.ResponseWriter, r *http.Request) {

	type profile struct {
		Name string
		Href string
		Desc string
	}
	var profiles []profile

	// Adding other profiles exposed from within this package
	for p, pd := range profileDescriptions {
		profiles = append(profiles, profile{
			Name: p,
			Href: p,
			Desc: pd,
		})
	}

	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	indexTmpl.Execute(w, profiles)
}

var profileDescriptions = map[string]string{

	"/logging": `View \ change the log level of the program. <br/>

log level:
<select id="loglevelselector">
<option value="debug">debug</option>
<option value="info">info</option>
<option value="warn">warn</option>
<option value="error">error</option>
</select>
<button onclick="setlevel(document.getElementById('loglevelselector').value)">click</button>

<script>
function setlevel(l) {
	var xhr = new XMLHttpRequest();
	xhr.open('PUT', '/logging', true);
	xhr.setRequestHeader("Content-Type", "application/json");

	xhr.onreadystatechange = function() {
		if (this.readyState == 4 && this.status == 200) {
			var resp = JSON.parse(this.responseText);
			alert("log level set to:" + resp["level"]);
		}
	};

	xhr.send('{"level":"' + l + '"}');
}
</script>
	`,
}

var indexTmpl = template.Must(template.New("index").Parse(`<!DOCTYPE html><html>
<head>
<title>/debug/pprof/</title>
<style>
.profile-name{
	display:inline-block;
	width:6rem;
}
</style>
</head>
<body>
Things to do:
{{range .}}
<h2><a href={{.Href}}>{{.Name}}</a></h2>
<p>
{{.Desc}}
</p>
{{end}}
</body>
</html>
`))
