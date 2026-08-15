package main

type Application struct {

}

func main(){
	app := &Application{}
	router := app.Routes()
	router.Listen(":3000")
}