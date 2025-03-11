package main

import (
	"iemhun/algorithms"
	"testing"
)

func BenchmarkRunEMHUN(b *testing.B) {
	fileName := "data/foodmart_dynamic.txt"
	minUtility := 20000.0

	// Đọc các giao dịch từ file một lần
	transactions, err := readTransactionsFromFile(fileName)
	if err != nil {
		b.Fatal(err)
	}

	// Khởi tạo EMHUN với các giao dịch đã đọc
	emhun := algorithms.NewEMHUN(transactions, minUtility)

	// Đặt lại bộ đếm thời gian để chỉ đo phần Run()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		emhun.Run()
	}
}
