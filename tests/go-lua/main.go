package main

func main() {
	L, err := NewLuaState()
	if err != nil {
		panic(err)
	}
	defer L.Close()
	L.OpenLibs()

	script := `
print(os.date())
`
	if err = L.DoString(script); err != nil {
		panic(err)
	}
}
