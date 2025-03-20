package main

import (
	"MessengerMPM/internal/algs"
	"bytes"
	"fmt"
	"log"
)

func main() {
	// Создаем экземпляр RC5
	cipher := algs.NewRC5()

	// Генерация случайного ключа с использованием RC5
	key := make([]byte, 16) // 128-битный ключ
	roundKeys, err := cipher.GenerateKeys(key)
	if err != nil {
		log.Fatalf("Ошибка генерации ключа: %v", err)
	}
	fmt.Printf("Сгенерированные раундовые ключи: %x\n", roundKeys)

	// Устанавливаем ключ
	if err := cipher.SetKey(key); err != nil {
		log.Fatalf("Ошибка установки ключа: %v", err)
	}

	// Создаем контекст с алгоритмом RC5
	context, err := algs.NewCryptoSymmetricContext(
		key,        // Ключ
		cipher,     // Алгоритм RC5
		algs.ECB,   // Режим шифрования
		algs.Zeros, // Режим набивки
		nil,        // IV (не требуется для ECB)
		16,         // Размер блока (128 бит)
	)
	if err != nil {
		log.Fatalf("Ошибка создания контекста: %v", err)
	}

	// Тестовые данные
	data := []byte("RC5 тестовые данные")
	fmt.Printf("Исходные данные: %s\n", data)

	// Добавляем набивку через метод контекста
	paddedData, err := context.AddPadding(data)
	if err != nil {
		log.Fatalf("Ошибка добавления набивки: %v", err)
	}

	// Шифрование
	encrypted, err := context.Encrypt(paddedData)
	if err != nil {
		log.Fatalf("Ошибка шифрования: %v", err)
	}
	fmt.Printf("Зашифрованные данные: %x\n", encrypted)

	// Дешифрование
	decrypted, err := context.Decrypt(encrypted)
	if err != nil {
		log.Fatalf("Ошибка дешифрования: %v", err)
	}

	// Удаляем набивку через метод контекста
	unpaddedData, err := context.RemovePadding(decrypted)
	if err != nil {
		log.Fatalf("Ошибка удаления набивки: %v", err)
	}
	fmt.Printf("Расшифрованные данные: %s\n", unpaddedData)

	// Проверка совпадения исходного текста с расшифрованным
	if bytes.Equal(data, unpaddedData) {
		fmt.Println("Шифрование и дешифрование прошли успешно!")
	} else {
		fmt.Println("Ошибка: данные после дешифрования не совпадают с исходными!")
	}
}
