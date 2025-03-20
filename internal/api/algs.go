package api

import (
	"MessengerMPM/internal/algs"
	"fmt"
	"log"
)

func EncryptRC5(data []byte, key []byte) ([]byte, error) {
	cipher := algs.NewRC5()

	roundKeys, err := cipher.GenerateKeys(key)
	if err != nil {
		log.Fatalf("Ошибка генерации ключа: %v", err)
	}
	log.Printf("Сгенерированные раундовые ключи: %x\n", roundKeys)

	if err := cipher.SetKey(key); err != nil {
		log.Printf("error setting key: %v", err)
		return nil, fmt.Errorf("error setting key: %v", err)
	}

	context, err := algs.NewCryptoSymmetricContext(
		key,
		cipher,
		algs.ECB,
		algs.Zeros,
		nil,
		16,
	)
	if err != nil {
		log.Printf("error creating context: %v", err)
		return nil, fmt.Errorf("error creating context: %v", err)
	}

	paddedData, err := context.AddPadding(data)
	if err != nil {
		log.Printf("error adding padding: %v", err)
		return nil, fmt.Errorf("error adding padding: %v", err)
	}

	encrypted, err := context.Encrypt(paddedData)
	if err != nil {
		log.Printf("error encrypting: %v", err)
		return nil, fmt.Errorf("error encrypting: %v", err)
	}

	return encrypted, nil
}

func DecryptRC5(data []byte, key []byte) ([]byte, error) {
	cipher := algs.NewRC5()

	if err := cipher.SetKey(key); err != nil {
		log.Printf("error: %v", err)
		return nil, fmt.Errorf("error setting key: %v", err)
	}

	context, err := algs.NewCryptoSymmetricContext(
		key,
		cipher,
		algs.ECB,
		algs.Zeros,
		nil,
		16,
	)
	if err != nil {
		log.Printf("creating context: %v", err)
		return nil, fmt.Errorf("error creating context: %v", err)
	}

	decrypted, err := context.Decrypt(data)
	if err != nil {
		log.Printf("error decrypting: %v", err)
		return nil, fmt.Errorf("error decrypting: %v", err)
	}

	unpaddedData, err := context.RemovePadding(decrypted)
	if err != nil {
		log.Printf("error removing padding: %v", err)
		return nil, fmt.Errorf("error removing padding: %v", err)
	}

	return unpaddedData, nil
}

func EncryptSerpent(data []byte, key []byte) ([]byte, error) {
	cipher := algs.NewSerpentCipher()

	if err := cipher.SetKey(key); err != nil {
		log.Printf("error: %v", err)
		return nil, fmt.Errorf("error setting key: %v", err)
	}

	context, err := algs.NewCryptoSymmetricContext(
		key,
		cipher,
		algs.ECB,
		algs.Zeros,
		nil,
		16,
	)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, fmt.Errorf("error creating context: %v", err)
	}

	paddedData, err := context.AddPadding(data)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, fmt.Errorf("error adding padding: %v", err)
	}

	encrypted, err := context.Encrypt(paddedData)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, fmt.Errorf("error encrypting: %v", err)
	}

	return encrypted, nil
}

func DecryptSerpent(data []byte, key []byte) ([]byte, error) {
	cipher := algs.NewSerpentCipher()

	if err := cipher.SetKey(key); err != nil {
		log.Printf("error: %v", err)
		return nil, fmt.Errorf("error setting key: %v", err)
	}

	context, err := algs.NewCryptoSymmetricContext(
		key,
		cipher,
		algs.ECB,
		algs.Zeros,
		nil,
		16,
	)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, fmt.Errorf("error creating context: %v", err)
	}

	decrypted, err := context.Decrypt(data)

	if err != nil {
		log.Printf("error: %v", err)
		return nil, fmt.Errorf("error decrypting: %v", err)
	}

	unpaddedData, err := context.RemovePadding(decrypted)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, fmt.Errorf("error removing padding: %v", err)
	}

	return unpaddedData, nil
}
