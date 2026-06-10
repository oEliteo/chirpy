package main

import (
	"testing"
)

func TestChirpFilterUnchanged(t *testing.T) {
	msgBody := "This chirp should be unchanged due to not containing banned words."
	result := filterChirp(msgBody, bannedWordsMap)

	if result != msgBody {
		t.Errorf("Got %q, expected %q", result, msgBody)
	}
}

func TestChirpFilterUnchanged2(t *testing.T) {
	msgBody := "This chirp should also be unchanged despite containing a banned word due to it having punctuation attached. Sharbert!"
	result := filterChirp(msgBody, bannedWordsMap)

	if result != msgBody {
		t.Errorf("Got %q, expected %q", result, msgBody)
	}
}

func TestChirpFilterOneBannedWord(t *testing.T) {
	msgBody := "These kerfuffle jerks stole my pudding cups."
	result := filterChirp(msgBody, bannedWordsMap)
	expected := "These **** jerks stole my pudding cups."

	if result != expected {
		t.Errorf("Got %q, expected %q", result, expected)
	}
}

func TestChirpFilterTwoBannedWordsOnePunctuated(t *testing.T) {
	msgBody := "I'm so fornax tired of hearing about sharbert. KERFUFFLE"
	result := filterChirp(msgBody, bannedWordsMap)
	expected := "I'm so **** tired of hearing about sharbert. ****"

	if result != expected {
		t.Errorf("Got %q, expected %q", result, expected)
	}
}

func TestChirpFilterThreeBannedWordsNoPunctuation(t *testing.T) {
	msgBody := "I'm so fornax sharbert kerfuffle man."
	result := filterChirp(msgBody, bannedWordsMap)
	expected := "I'm so **** **** **** man."

	if result != expected {
		t.Errorf("Got %q, expected %q", result, expected)
	}
}
