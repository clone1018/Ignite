package controllers

import (
	"fmt"
	"github.com/robfig/revel"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type App struct {
	*revel.Controller
}

type Program struct {
	Language string
	Code     string
}

type CompileResult struct {
	Output string
}

func (c App) Index() revel.Result {
	return c.Render()
}

func (c App) Run(code string, language string) revel.Result {
	program := Program{Language: language, Code: code}

	output := RunCode(program)
	result := CompileResult{Output: output}

	return c.RenderJson(result)
}

func CreateProgram(contents string) string {
	rand.Seed(time.Now().UTC().UnixNano())
	name := randomString(10)

	os.Mkdir(filepath.Join("/tmp", name), os.ModeDir)

	d1 := []byte(contents)
	err := ioutil.WriteFile(filepath.Join("/tmp", name, name+".php"), d1, 0644)
	if err != nil {
		panic(err)
	}

	return name
}

func RunCode(program Program) string {
	name := CreateProgram(program.Code)

	cmd := exec.Command("/usr/bin/docker", "run", "-v", "/tmp/"+name+":/tmp/"+name, "prehash/php5", "php", "-f", "/tmp/"+name+"/"+name+".php")
	err := cmd.Start()
	if err != nil {
		panic(err)
	}

	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-time.After(3 * time.Second):
		if err := cmd.Process.Kill(); err != nil {
			panic("failed to kill")
		}
		<-done // allow goroutine to exit
		return fmt.Sprint("Process killed.")
	case err := <-done:
		return fmt.Sprintf("%v.", err)
	}

	out, _ := cmd.Output()

	return fmt.Sprintf(string(out[:]))
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
