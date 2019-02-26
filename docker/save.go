package docker

// Save saves image to dest, as in `docker save`
func Save(image, dest string) error {
	return Command("save", "-o", dest, image).Run()
}
