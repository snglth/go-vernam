package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type void struct{}

var item void

type VernKey struct {
	alphabet []rune
	key      string
}

func (vernKey *VernKey) Validate() error {
	alphRunes := make(map[rune]void)
	for _, alphRune := range vernKey.alphabet {
		alphRunes[alphRune] = item
	}

	for _, keyMember := range vernKey.key {
		if _, ok := alphRunes[keyMember]; !ok {
			return errors.New(fmt.Sprintf("key member '%s' is out of the alphabet", string(keyMember)))
		}
	}

	return nil
}

func (vernKey *VernKey) RuneToIndex(rl rune) (int, error) {
	for rune_i, letter := range vernKey.alphabet {
		if rl == letter {
			return rune_i, nil
		}
	}

	return -1, errors.New(fmt.Sprintf("letter '%s' is out of alphabet", string(rl)))
}

func (vernKey *VernKey) IndexToRune(index int) rune {
	return vernKey.alphabet[index]
}

func ReadKey(keyFile string) (*VernKey, error) {
	key_json, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}

	var keyMap map[string]interface{}
	err = json.Unmarshal([]byte(key_json), &keyMap)
	if err != nil {
		return nil, err
	}

	alphabet := keyMap["alphabet"].([]interface{})
	alphLen := len(alphabet)
	alphRunes := make([]rune, alphLen, alphLen)
	for alph_i, alphItem := range alphabet {
		var alphRune rune
		for _, c := range alphItem.(string) {
			alphRune = c
			break
		}
		alphRunes[alph_i] = alphRune
	}

	key := keyMap["key"].(string)

	vernKey := VernKey{alphRunes, key}
	return &vernKey, nil
}

func EncodeStream(reader *bufio.Reader, writer *bufio.Writer, vernKey *VernKey) error {
	keyLen := len(vernKey.key)
	indexKey := make([]int, keyLen, keyLen)
	for keyIndex, keyRune := range vernKey.key {
		index, _ := vernKey.RuneToIndex(keyRune)
		indexKey[keyIndex] = index
	}

	for {
		rune, _, error := reader.ReadRune()
		if error == io.EOF || rune == '\n' {
			break
		}

		runeIndex, err := vernKey.RuneToIndex(rune)
		if err != nil {
			return error
		}

		for _, index := range indexKey {
			runeIndex ^= index
		}

		writer.WriteRune(vernKey.IndexToRune(runeIndex))
	}

	writer.WriteRune('\n')
	writer.Flush()
	return nil
}

func DecodeStream(reader *bufio.Reader, writer *bufio.Writer, vernKey *VernKey) error {
	keyLen := len(vernKey.key)
	indexKey := make([]int, keyLen, keyLen)
	for keyIndex, keyRune := range vernKey.key {
		index, _ := vernKey.RuneToIndex(keyRune)
		indexKey[keyIndex] = index
	}

	for i, j := 0, len(indexKey)-1; i < j; i, j = i+1, j-1 {
		indexKey[i], indexKey[j] = indexKey[j], indexKey[i]
	}

	for {
		rune, _, error := reader.ReadRune()
		if error == io.EOF || rune == '\n' {
			break
		}

		runeIndex, err := vernKey.RuneToIndex(rune)
		if err != nil {
			return error
		}

		for _, index := range indexKey {
			runeIndex ^= index
		}

		writer.WriteRune(vernKey.IndexToRune(runeIndex))
	}

	writer.WriteRune('\n')
	writer.Flush()
	return nil
}

func main() {
	keyFilePtr := flag.String("key", "key.json", "key file")
	decodePtr := flag.Bool("decode", false, "decode mode")

	flag.Parse()

	vernKey, err := ReadKey(*keyFilePtr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = vernKey.Validate()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	if *decodePtr {
		err = DecodeStream(reader, writer, vernKey)
	} else {
		err = EncodeStream(reader, writer, vernKey)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
