package main

import forum "forum/src"

func main() {
	forum.CreateDB() //ATT db avant server sinon se lance PAS
	forum.StartServer()
}
