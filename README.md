swtpm_test
==========

Small library to setup a user-accessible swtpm instance.

Usefull for writing test-suites that involves TPMs.

```go
func main() {
	dir, err := os.MkdirTemp("/var/tmp", "example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up

	tpm := swtpm.NewSwtpm(dir)
	socket, err := tpm.Socket()
	if err != nil {
		log.Fatal("failed socket", err)
	}
	defer tpm.Stop()

	if _, err := tpm2.OpenTPM(socket); err != nil {
		log.Fatal(err)
	}
}
```
