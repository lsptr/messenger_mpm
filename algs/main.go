package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"log"

	"MessengerMPM/algs/serpent"
)

func main() {
	// Генерация случайного ключа
	key := make([]byte, 32) // 256-битный ключ
	if _, err := rand.Read(key); err != nil {
		log.Fatalf("Ошибка генерации ключа: %v", err)
	}

	// Создаем экземпляр Serpent
	cipher := serpent.NewSerpentCipher()

	// Устанавливаем ключ
	if err := cipher.SetKey(key); err != nil {
		log.Fatalf("Ошибка установки ключа: %v", err)
	}

	// Создаем данные для тестирования
	data := []byte("Тестовые данные Serpent!")
	data = append(data, bytes.Repeat([]byte{0}, 16-len(data)%16)...) // Дополняем до кратного 16

	// Шифрование
	encrypted, err := cipher.Encrypt(data)
	if err != nil {
		log.Fatalf("Ошибка шифрования: %v", err)
	}
	fmt.Printf("Зашифрованные данные: %x\n", encrypted)

	// Дешифрование
	decrypted, err := cipher.Decrypt(encrypted)
	if err != nil {
		log.Fatalf("Ошибка дешифрования: %v", err)
	}
	fmt.Printf("Расшифрованные данные: %s\n", decrypted)

	// Проверка совпадения исходного текста с расшифрованным
	if bytes.Equal(data, decrypted) {
		fmt.Println("Шифрование и дешифрование прошли успешно!")
	} else {
		fmt.Println("Ошибка: расшифрованные данные не совпадают с исходными!")
	}
}
