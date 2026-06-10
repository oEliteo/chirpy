package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChirpFilterUnchanged(t *testing.T) {
	msgBody := "This chirp should be unchanged due to not containing banned words."
	result := filterChirp(msgBody, bannedWordsMap)

	if result != msgBody {
		t.Errorf("Got %s, expected %s", result, msgBody)
	}
}

func TestChirpFilterUnchanged2(t *testing.T) {
	msgBody := "This chirp should also be unchanged despite containing a banned word due to it having punctuation attached. Sharbert!"
	result := filterChirp(msgBody, bannedWordsMap)

	if result != msgBody {
		t.Errorf("Got %s, expected %s", result, msgBody)
	}
}

func TestChirpFilterOneBannedWord(t *testing.T) {
	msgBody := "These kerfuffle jerks stole my pudding cups."
	result := filterChirp(msgBody, bannedWordsMap)
	expected := "These **** jerks stole my pudding cups."

	if result != expected {
		t.Errorf("Got %s, expected %s", result, expected)
	}
}

func TestChirpFilterTwoBannedWordsOnePunctuated(t *testing.T) {
	msgBody := "I'm so fornax tired of hearing about sharbert. KERFUFFLE"
	result := filterChirp(msgBody, bannedWordsMap)
	expected := "I'm so **** tired of hearing about sharbert. ****"

	if result != expected {
		t.Errorf("Got %s, expected %s", result, expected)
	}
}

func TestChirpFilterThreeBannedWordsNoPunctuation(t *testing.T) {
	msgBody := "I'm so fornax sharbert kerfuffle man."
	result := filterChirp(msgBody, bannedWordsMap)
	expected := "I'm so **** **** **** man."

	if result != expected {
		t.Errorf("Got %s, expected %s", result, expected)
	}
}

func TestValidateChirpTooLong(t *testing.T) {
	body := strings.NewReader(`{"body":"` + strings.Repeat("a", 141) + `"}`)
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/validate_chirp", body)
	rec := httptest.NewRecorder()

	cfg := apiConfig{}
	cfg.validateChirp(rec, req)
	res := rec.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d, got %d", http.StatusBadRequest, res.StatusCode)
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Could not read response body: %v", err)
	}
	bodyString := string(resBody)

	expected := `{"error":"Chirp is too long"}`
	if bodyString != expected {
		t.Errorf("Got %s, expected %s", bodyString, expected)
	}
}

func TestValidateChirpMaxLength(t *testing.T) {
	body := strings.NewReader(`{"body":"` + strings.Repeat("a", 140) + `"}`)
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/validate_chirp", body)
	rec := httptest.NewRecorder()

	cfg := apiConfig{}
	cfg.validateChirp(rec, req)
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Got %d, expected %d", res.StatusCode, http.StatusOK)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Could not read response body: %s", err)
	}
	bodyString := string(resBody)

	expected := `{"cleaned_body":"` + strings.Repeat("a", 140) + `"}`
	if bodyString != expected {
		t.Errorf("Got %s, expected %s", bodyString, expected)
	}
}
