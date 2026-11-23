package password

import "testing"

func TestHashAndVerify(t *testing.T) {
	hash, err := Hash("student12345")
	if err != nil {
		t.Fatal(err)
	}
	ok, err := Verify("student12345", hash)
	if err != nil || !ok {
		t.Fatalf("пароль должен подходить: ok=%v err=%v", ok, err)
	}
	ok, err = Verify("wrong-password", hash)
	if err != nil || ok {
		t.Fatalf("чужой пароль не должен подходить: ok=%v err=%v", ok, err)
	}
	if _, err := Verify("student12345", "not-a-hash"); err == nil {
		t.Fatal("битый хеш должен возвращать ошибку")
	}
}
