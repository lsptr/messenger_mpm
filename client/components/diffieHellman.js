import bigInt from 'big-integer';

// Генерация приватного ключа
export function generatePrivateKey(p) {
    return bigInt.randBetween(1, bigInt(p).minus(1));
}

// Генерация публичного ключа (g^privateKey mod p)
export function generatePublicKey(g, privateKey, p) {
    return bigInt(g).pow(privateKey).mod(p);
}

// Вычисление общего секретного ключа (publicKey^privateKey mod p)
export function computeSharedSecret(publicKey, privateKey, p) {
    return bigInt(publicKey).pow(privateKey).mod(p);
}