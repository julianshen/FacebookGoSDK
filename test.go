package main

import (
	"./facebook"
	"fmt"
)

var s string = "dM62xGl1rG84sTb7GCkGzqdi_sUslKhjIpF0DfL0GK4.eyJhbGdvcml0aG0iOiJITUFDLVNIQTI1NiIsImNvZGUiOiIyLkFRQlRLOVNteW1QYUl3cTAuMzYwMC4xMzUxMzQ2NDAwLjEtMTEyOTI4MzQzN3wxMzUxMzQwNTg1fGl3bko2SG9sb1pTY0Vvb1IxR2R3aGhZVmFQMCIsImlzc3VlZF9hdCI6MTM1MTM0MDI4NSwidXNlcl9pZCI6IjExMjkyODM0MzcifQ"

func main() {

	f := facebook.NewBasicContext("111027248936016", "b72d2afbaf8f70c6206b3ad9c52eff27")

	f.ParseSignedRequest(s)

	fmt.Println(f.AccessToken())
}
