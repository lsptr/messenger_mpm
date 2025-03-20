package serpent

import (
	"errors"
)

const BlockSize = 16

const phi = 0x9e3779b9

var errKeySize = errors.New("invalid key size")
var errBlockSize = errors.New("invalid block size")

// SerpentCipher реализует интерфейсы KeyRound, CipherTransform и SymmetricAlgorithm
type SerpentCipher struct {
	roundKeys [132]uint32
}

// NewSerpentCipher создает новый экземпляр алгоритма Serpent
func NewSerpentCipher() *SerpentCipher {
	return &SerpentCipher{}
}

// GenerateKeys генерирует раундовые ключи из входного ключа
func (s *SerpentCipher) GenerateKeys(key []byte) ([][]byte, error) {
	if k := len(key); k != 16 && k != 24 && k != 32 {
		return nil, errKeySize
	}

	s.keySchedule(key)

	// Преобразуем раундовые ключи в формат [][]byte
	roundKeys := make([][]byte, 33) // 32 раунда + whitening
	for i := 0; i < 33; i++ {
		roundKeys[i] = make([]byte, 16)
		for j := 0; j < 4; j++ {
			idx := i*4 + j
			roundKeys[i][j*4] = byte(s.roundKeys[idx])
			roundKeys[i][j*4+1] = byte(s.roundKeys[idx] >> 8)
			roundKeys[i][j*4+2] = byte(s.roundKeys[idx] >> 16)
			roundKeys[i][j*4+3] = byte(s.roundKeys[idx] >> 24)
		}
	}

	return roundKeys, nil
}

// Encryption шифрует один блок данных
func (s *SerpentCipher) Encryption(block, roundKey []byte) ([]byte, error) {
	if len(block) != BlockSize {
		return nil, errBlockSize
	}
	if len(roundKey) != BlockSize {
		return nil, errBlockSize
	}

	dst := make([]byte, BlockSize)

	// Преобразуем раундовый ключ в uint32
	var sk [4]uint32
	for i := 0; i < 4; i++ {
		sk[i] = uint32(roundKey[i*4]) | uint32(roundKey[i*4+1])<<8 | uint32(roundKey[i*4+2])<<16 | uint32(roundKey[i*4+3])<<24
	}

	// Преобразуем входной блок в 4 x 32 bit регистры
	r0 := uint32(block[0]) | uint32(block[1])<<8 | uint32(block[2])<<16 | uint32(block[3])<<24
	r1 := uint32(block[4]) | uint32(block[5])<<8 | uint32(block[6])<<16 | uint32(block[7])<<24
	r2 := uint32(block[8]) | uint32(block[9])<<8 | uint32(block[10])<<16 | uint32(block[11])<<24
	r3 := uint32(block[12]) | uint32(block[13])<<8 | uint32(block[14])<<16 | uint32(block[15])<<24

	// XOR с раундовым ключом
	r0 ^= sk[0]
	r1 ^= sk[1]
	r2 ^= sk[2]
	r3 ^= sk[3]

	// Применяем S-блок и линейное преобразование
	sb0(&r0, &r1, &r2, &r3)
	linear(&r0, &r1, &r2, &r3)

	// Записываем результат
	dst[0] = byte(r0)
	dst[1] = byte(r0 >> 8)
	dst[2] = byte(r0 >> 16)
	dst[3] = byte(r0 >> 24)
	dst[4] = byte(r1)
	dst[5] = byte(r1 >> 8)
	dst[6] = byte(r1 >> 16)
	dst[7] = byte(r1 >> 24)
	dst[8] = byte(r2)
	dst[9] = byte(r2 >> 8)
	dst[10] = byte(r2 >> 16)
	dst[11] = byte(r2 >> 24)
	dst[12] = byte(r3)
	dst[13] = byte(r3 >> 8)
	dst[14] = byte(r3 >> 16)
	dst[15] = byte(r3 >> 24)

	return dst, nil
}

// Decryption расшифровывает один блок данных
func (s *SerpentCipher) Decryption(block, roundKey []byte) ([]byte, error) {
	if len(block) != BlockSize {
		return nil, errBlockSize
	}
	if len(roundKey) != BlockSize {
		return nil, errBlockSize
	}

	dst := make([]byte, BlockSize)

	// Преобразуем раундовый ключ в uint32
	var sk [4]uint32
	for i := 0; i < 4; i++ {
		sk[i] = uint32(roundKey[i*4]) | uint32(roundKey[i*4+1])<<8 | uint32(roundKey[i*4+2])<<16 | uint32(roundKey[i*4+3])<<24
	}

	// Преобразуем входной блок в 4 x 32 bit регистры
	r0 := uint32(block[0]) | uint32(block[1])<<8 | uint32(block[2])<<16 | uint32(block[3])<<24
	r1 := uint32(block[4]) | uint32(block[5])<<8 | uint32(block[6])<<16 | uint32(block[7])<<24
	r2 := uint32(block[8]) | uint32(block[9])<<8 | uint32(block[10])<<16 | uint32(block[11])<<24
	r3 := uint32(block[12]) | uint32(block[13])<<8 | uint32(block[14])<<16 | uint32(block[15])<<24

	// Применяем обратное линейное преобразование
	linearInv(&r0, &r1, &r2, &r3)

	// Применяем обратный S-блок
	sb0Inv(&r0, &r1, &r2, &r3)

	// XOR с раундовым ключом
	r0 ^= sk[0]
	r1 ^= sk[1]
	r2 ^= sk[2]
	r3 ^= sk[3]

	// Записываем результат
	dst[0] = byte(r0)
	dst[1] = byte(r0 >> 8)
	dst[2] = byte(r0 >> 16)
	dst[3] = byte(r0 >> 24)
	dst[4] = byte(r1)
	dst[5] = byte(r1 >> 8)
	dst[6] = byte(r1 >> 16)
	dst[7] = byte(r1 >> 24)
	dst[8] = byte(r2)
	dst[9] = byte(r2 >> 8)
	dst[10] = byte(r2 >> 16)
	dst[11] = byte(r2 >> 24)
	dst[12] = byte(r3)
	dst[13] = byte(r3 >> 8)
	dst[14] = byte(r3 >> 16)
	dst[15] = byte(r3 >> 24)

	return dst, nil
}

// SetKey устанавливает ключ шифрования
func (s *SerpentCipher) SetKey(key []byte) error {
	if k := len(key); k != 16 && k != 24 && k != 32 {
		return errKeySize
	}
	s.keySchedule(key)
	return nil
}

// Encrypt шифрует блок данных
func (s *SerpentCipher) Encrypt(data []byte) ([]byte, error) {
	if len(data) != BlockSize {
		return nil, errBlockSize
	}

	dst := make([]byte, BlockSize)
	encryptBlock(dst, data, &s.roundKeys)
	return dst, nil
}

// Decrypt расшифровывает блок данных
func (s *SerpentCipher) Decrypt(data []byte) ([]byte, error) {
	if len(data) != BlockSize {
		return nil, errBlockSize
	}

	dst := make([]byte, BlockSize)
	decryptBlock(dst, data, &s.roundKeys)
	return dst, nil
}

func (s *SerpentCipher) keySchedule(key []byte) {
	var k [16]uint32
	j := 0
	for i := 0; i+4 <= len(key); i += 4 {
		k[j] = uint32(key[i]) | uint32(key[i+1])<<8 | uint32(key[i+2])<<16 | uint32(key[i+3])<<24
		j++
	}
	if j < 8 {
		k[j] = 1
	}

	for i := 8; i < 16; i++ {
		x := k[i-8] ^ k[i-5] ^ k[i-3] ^ k[i-1] ^ phi ^ uint32(i-8)
		k[i] = (x << 11) | (x >> 21)
		s.roundKeys[i-8] = k[i]
	}
	for i := 8; i < 132; i++ {
		x := s.roundKeys[i-8] ^ s.roundKeys[i-5] ^ s.roundKeys[i-3] ^ s.roundKeys[i-1] ^ phi ^ uint32(i)
		s.roundKeys[i] = (x << 11) | (x >> 21)
	}

	sb3(&s.roundKeys[0], &s.roundKeys[1], &s.roundKeys[2], &s.roundKeys[3])
	sb2(&s.roundKeys[4], &s.roundKeys[5], &s.roundKeys[6], &s.roundKeys[7])
	sb1(&s.roundKeys[8], &s.roundKeys[9], &s.roundKeys[10], &s.roundKeys[11])
	sb0(&s.roundKeys[12], &s.roundKeys[13], &s.roundKeys[14], &s.roundKeys[15])
	sb7(&s.roundKeys[16], &s.roundKeys[17], &s.roundKeys[18], &s.roundKeys[19])
	sb6(&s.roundKeys[20], &s.roundKeys[21], &s.roundKeys[22], &s.roundKeys[23])
	sb5(&s.roundKeys[24], &s.roundKeys[25], &s.roundKeys[26], &s.roundKeys[27])
	sb4(&s.roundKeys[28], &s.roundKeys[29], &s.roundKeys[30], &s.roundKeys[31])

	sb3(&s.roundKeys[32], &s.roundKeys[33], &s.roundKeys[34], &s.roundKeys[35])
	sb2(&s.roundKeys[36], &s.roundKeys[37], &s.roundKeys[38], &s.roundKeys[39])
	sb1(&s.roundKeys[40], &s.roundKeys[41], &s.roundKeys[42], &s.roundKeys[43])
	sb0(&s.roundKeys[44], &s.roundKeys[45], &s.roundKeys[46], &s.roundKeys[47])
	sb7(&s.roundKeys[48], &s.roundKeys[49], &s.roundKeys[50], &s.roundKeys[51])
	sb6(&s.roundKeys[52], &s.roundKeys[53], &s.roundKeys[54], &s.roundKeys[55])
	sb5(&s.roundKeys[56], &s.roundKeys[57], &s.roundKeys[58], &s.roundKeys[59])
	sb4(&s.roundKeys[60], &s.roundKeys[61], &s.roundKeys[62], &s.roundKeys[63])

	sb3(&s.roundKeys[64], &s.roundKeys[65], &s.roundKeys[66], &s.roundKeys[67])
	sb2(&s.roundKeys[68], &s.roundKeys[69], &s.roundKeys[70], &s.roundKeys[71])
	sb1(&s.roundKeys[72], &s.roundKeys[73], &s.roundKeys[74], &s.roundKeys[75])
	sb0(&s.roundKeys[76], &s.roundKeys[77], &s.roundKeys[78], &s.roundKeys[79])
	sb7(&s.roundKeys[80], &s.roundKeys[81], &s.roundKeys[82], &s.roundKeys[83])
	sb6(&s.roundKeys[84], &s.roundKeys[85], &s.roundKeys[86], &s.roundKeys[87])
	sb5(&s.roundKeys[88], &s.roundKeys[89], &s.roundKeys[90], &s.roundKeys[91])
	sb4(&s.roundKeys[92], &s.roundKeys[93], &s.roundKeys[94], &s.roundKeys[95])

	sb3(&s.roundKeys[96], &s.roundKeys[97], &s.roundKeys[98], &s.roundKeys[99])
	sb2(&s.roundKeys[100], &s.roundKeys[101], &s.roundKeys[102], &s.roundKeys[103])
	sb1(&s.roundKeys[104], &s.roundKeys[105], &s.roundKeys[106], &s.roundKeys[107])
	sb0(&s.roundKeys[108], &s.roundKeys[109], &s.roundKeys[110], &s.roundKeys[111])
	sb7(&s.roundKeys[112], &s.roundKeys[113], &s.roundKeys[114], &s.roundKeys[115])
	sb6(&s.roundKeys[116], &s.roundKeys[117], &s.roundKeys[118], &s.roundKeys[119])
	sb5(&s.roundKeys[120], &s.roundKeys[121], &s.roundKeys[122], &s.roundKeys[123])
	sb4(&s.roundKeys[124], &s.roundKeys[125], &s.roundKeys[126], &s.roundKeys[127])

	sb3(&s.roundKeys[128], &s.roundKeys[129], &s.roundKeys[130], &s.roundKeys[131])
}
