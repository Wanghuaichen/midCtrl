// Package main 程序的主体和入口

package main

func initLoger() error {
	file, err := os.OpenFile("./log.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModeAppend)
	if err != nil {
		return err
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.SetOutput(file)
		return nil
	}
}

func main(){
}
