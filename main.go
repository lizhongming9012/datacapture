package main

var c = Conf{}

func init() {
	c.ConfReader()
	SetupDataBase()
}

func main() {
	StartZsk()
}
