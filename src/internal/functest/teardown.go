package functest

// TearDown cleans up any resources allocated during the test.
func TearDown() {
	tearDownNamespaces()
	tearDownPeers()
}
