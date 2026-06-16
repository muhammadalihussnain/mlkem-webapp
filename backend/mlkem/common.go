package mlkem

const Q int32 = 3329

// barrettReduce reduces a 32-bit integer modulo Q using Barrett reduction
func barrettReduce(x int32) int32 {
    const v = 20159 // floor(2^26 / 3329) + 1
    t := (int64(x) * v) >> 26
    x = int32(int64(x) - t*int64(Q))
    if x >= Q {
        x -= Q
    }
    if x < 0 {
        x += Q
    }
    return x
}

// mulMontgomery performs multiplication modulo Q
func mulMontgomery(a, b int32) int32 {
    return barrettReduce(a * b)
}
