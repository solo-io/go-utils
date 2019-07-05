This project contains some utilities for GCloud, most importantly a library for building google cloud subscriber 
plugins for Solobot CI builds. 

To build a new cloud subscriber, simply do:

```
subscriber, err := NewSolobotCloudSubscriber(ctx context.Context)
if err != nil {
  // handle
}
subscriber.RegisterHandler(myCustomHandler)
if err := subscriber.Run(); err != nil {
  // handle
}
```

