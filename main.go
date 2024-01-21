package main

func main() {
	if ok, fd := calledAsPlugin(); ok {
		plugin(fd)
	}

	standalone()
}
