package main

import (
	"bufio"
	"fmt"
	"iemhun/algorithms"
	"iemhun/models"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func main() {
	fileName := "data/table3.txt"
	minUtility := 25.0

	// Đo thời gian bắt đầu
	startTime := time.Now()

	// Đo bộ nhớ trước khi chạy thuật toán
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	transactions, err := readTransactionsFromFile(fileName)
	if err != nil {
		fmt.Println("Error reading transactions:", err)
		return
	}
	fmt.Println("Transactions vừa đọc được:")
	for i, transaction := range transactions {
		fmt.Printf("Transaction %d: %s\n", i+1, transaction)
	}
	emhun := algorithms.NewEMHUN(transactions, minUtility)

	emhun.Run()

	elapsedTime := time.Since(startTime).Seconds()
	fmt.Printf("\nThời gian chạy thuật toán: %.6f s\n", elapsedTime)

	// Đo bộ nhớ sau khi chạy thuật toán
	runtime.ReadMemStats(&memStatsAfter)
	allocatedMemory := (memStatsAfter.Alloc - memStatsBefore.Alloc) / 1024
	fmt.Printf("Bộ nhớ sử dụng: %d KB\n", allocatedMemory)

	fmt.Println("\nFinished executing EMHUN algorithm.")
	outputFileName := "output/table3.txt"
	err = writeResultsToFile(emhun, outputFileName, elapsedTime, allocatedMemory)
	if err != nil {
		fmt.Println("Error writing results:", err)
		return
	}

	fmt.Println("Finished executing EMHUN algorithm. Results written to", outputFileName)
}

func readTransactionsFromFile(fileName string) ([]*models.Transaction, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var transactions []*models.Transaction

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) != 3 {
			fmt.Println("Invalid line format:", line)
			continue
		}

		itemsStr := strings.Fields(parts[0])
		var items []int
		for _, item := range itemsStr {
			itemInt, err := strconv.Atoi(item)
			if err != nil {
				return nil, err
			}
			items = append(items, itemInt)
		}

		transUtility, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return nil, err
		}

		utilitiesStr := strings.Fields(parts[2])
		var utilities []float64
		for _, utility := range utilitiesStr {
			utilityFloat, err := strconv.ParseFloat(utility, 64)
			if err != nil {
				return nil, err
			}
			utilities = append(utilities, utilityFloat)
		}

		// Tạo transaction với các số thực
		transaction := models.NewTransaction(items, utilities, transUtility)
		transactions = append(transactions, transaction)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}
func writeResultsToFile(emhun *algorithms.EMHUN, fileName string, elapsedTime float64, allocatedMemory uint64) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Ghi kết quả thuật toán
	for _, hui := range emhun.SearchAlgorithms.HighUtilityItemsets {
		line := fmt.Sprintf("Itemset: %v, Utility: %.2f\n", hui.Itemset, hui.Utility)
		_, err := writer.WriteString(line)
		if err != nil {
			return err
		}
	}

	// Ghi thông tin về thời gian (theo giây) và bộ nhớ
	_, err = writer.WriteString(fmt.Sprintf("\nThời gian chạy thuật toán: %.6f giây\n", elapsedTime))
	if err != nil {
		return err
	}

	_, err = writer.WriteString(fmt.Sprintf("Bộ nhớ sử dụng: %d KB\n", allocatedMemory))
	if err != nil {
		return err
	}

	return writer.Flush()
}
