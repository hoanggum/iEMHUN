package algorithms

import (
	"fmt"
	"iemhun/models"
	"iemhun/utility"
	"sort"
)

type EMHUN struct {
	Transactions       []*models.Transaction
	MinUtility         float64
	Rho, Delta, Eta    map[int]bool
	SortedSecondary    []int
	SortedEta          []int
	PrimaryItems       []int
	UtilityArray       *models.UtilityArray
	SearchAlgorithms   *SearchAlgorithms
	ItemTransactionMap map[int][]int
}

func NewEMHUN(transactions []*models.Transaction, minUtility float64) *EMHUN {
	utilityArray := models.NewUtilityArray()
	return &EMHUN{
		Transactions:       transactions,
		MinUtility:         minUtility,
		Rho:                make(map[int]bool),
		Delta:              make(map[int]bool),
		Eta:                make(map[int]bool),
		UtilityArray:       utilityArray,
		SearchAlgorithms:   NewSearchAlgorithms(utilityArray),
		ItemTransactionMap: make(map[int][]int),
	}
}

func (e *EMHUN) Run() {

	fmt.Println("Running EMHUN...")
	e.BuildItemTransactionMap()

	fmt.Println("Item Transaction Map:")
	e.PrintItemTransactionMap()
	e.ClassifyItems()

	fmt.Println("\nAfter classify, we have:")
	e.printClassification()

	fmt.Println("\nCalculating RTWU for all items in (ρ ∪ δ):")
	utility.CalculateRTWUForAllItems(e.Transactions, e.ItemTransactionMap, e.Rho, e.Delta, e.Eta, e.UtilityArray)

	fmt.Println("\nUA:")
	e.UtilityArray.PrintUtilityArray()

	combinedSet := e.unionKeys(e.Rho, e.Delta)
	secondaryItems := e.getSecondaryItems(combinedSet, e.UtilityArray, e.MinUtility)

	e.SortedSecondary = e.sortItems(secondaryItems)
	e.SortedEta = e.sortItems(e.keys(e.Eta))

	fmt.Println("\nSortedSecondary:", e.SortedSecondary)
	fmt.Println("\nSortedSecondary:", e.SortedEta)

	secondaryItemsMap := convertSliceToMap(e.SortedSecondary)
	e.FilterTransactions(secondaryItemsMap, e.Eta)

	e.PrintTransactions()
	e.PrintItemTransactionMap()

	e.SortItemsInTransactions()
	e.PrintTransactions()

	fmt.Println("\nSorting transactions by total RTWU:")
	e.SortTransactionsByTWU()
	fmt.Println("\nTransactions after sorting by RTWU:")
	e.PrintTransactions()
	fmt.Println("\nCalculating RSU for each item in Secondary(X)...")
	utility.CalculateRSUForAllItems(e.Transactions, e.ItemTransactionMap, e.SortedSecondary, e.UtilityArray)
	fmt.Println("\nUA:")
	e.UtilityArray.PrintUtilityArray()
	e.identifyPrimaryItems()
	fmt.Println("Primary:", e.PrimaryItems)
	fmt.Println("\nStarting HUI Search...")
	e.SearchAlgorithms.Search(e.SortedEta, make(map[int]bool), e.Transactions, e.PrimaryItems, e.SortedSecondary, e.MinUtility)

	// In kết quả sau khi tìm High Utility Itemsets
	fmt.Println("\nHUIs Found:")
	for _, hui := range e.SearchAlgorithms.HighUtilityItemsets {
		fmt.Printf("Itemset: %v, Utility: %.2f\n", hui.Itemset, hui.Utility)
	}
}

func (e *EMHUN) PrintTransactions() {
	fmt.Println("---------------------<Transaction>-------------------------")
	for i, transaction := range e.Transactions {
		fmt.Printf("Transaction %d: %s\n", i+1, transaction)
	}
	fmt.Println("-----------------------------------------------------------")
}
func (e *EMHUN) PrintItemTransactionMap() {
	fmt.Println("\nItem Transaction Map:")
	for item, transactions := range e.ItemTransactionMap {
		fmt.Printf("Item %d: %v\n", item, transactions)
	}
}

func (e *EMHUN) BuildItemTransactionMap() {
	for i, transaction := range e.Transactions {
		for _, item := range transaction.Items {
			e.ItemTransactionMap[item] = append(e.ItemTransactionMap[item], i+1) // Lưu vị trí transaction
		}
	}
}

func (e *EMHUN) ClassifyItems() {
	hasPositive := make(map[int]bool)
	hasNegative := make(map[int]bool)

	for _, transaction := range e.Transactions {
		for i, item := range transaction.Items {
			utility := transaction.Utilities[i]

			if utility > 0 {
				hasPositive[item] = true
			} else if utility < 0 {
				hasNegative[item] = true
			}
		}
	}

	allItems := e.unionKeys(hasPositive, hasNegative)

	for item := range allItems {
		positive := hasPositive[item]
		negative := hasNegative[item]

		if positive && !negative {
			e.Rho[item] = true
		} else if positive && negative {
			e.Delta[item] = true
		} else if negative && !positive {
			e.Eta[item] = true
		}
	}
}

func (e *EMHUN) printClassification() {
	rhoItems := e.keys(e.Rho)
	deltaItems := e.keys(e.Delta)
	etaItems := e.keys(e.Eta)

	sort.Ints(rhoItems)
	sort.Ints(deltaItems)
	sort.Ints(etaItems)

	fmt.Println("Items with positive utility only (ρ):", rhoItems)
	fmt.Println("Items with both positive and negative utility (δ):", deltaItems)
	fmt.Println("Items with negative utility only (η):", etaItems)
}

func (e *EMHUN) getSecondaryItems(combinedSet map[int]bool, utilityArray *models.UtilityArray, minU float64) []int {
	var secondary []int
	for item := range combinedSet {
		rlu := utilityArray.GetRTWU(item)
		if rlu >= minU {
			secondary = append(secondary, item)
		}
	}
	sort.Ints(secondary)
	fmt.Printf("Secondary(X) items: %v\n", secondary)
	return secondary
}

func (e *EMHUN) sortItems(items []int) []int {
	sort.Slice(items, func(i, j int) bool {
		typeOrderI := e.getTypeOrder(items[i])
		typeOrderJ := e.getTypeOrder(items[j])

		if typeOrderI != typeOrderJ {
			return typeOrderI < typeOrderJ
		}

		rtwuI := e.UtilityArray.GetRTWU(items[i])
		rtwuJ := e.UtilityArray.GetRTWU(items[j])

		return rtwuI < rtwuJ
	})

	return items
}

func (e *EMHUN) FilterTransactions(secondaryItems map[int]bool, etaItems map[int]bool) {
	// Xóa và tạo lại ItemTransactionMap
	e.ItemTransactionMap = make(map[int][]int)

	for index, transaction := range e.Transactions {
		var filteredItems []int
		var filteredUtilities []float64

		for i, item := range transaction.Items {
			if secondaryItems[item] || etaItems[item] {
				filteredItems = append(filteredItems, item)
				filteredUtilities = append(filteredUtilities, transaction.Utilities[i])

				// Cập nhật ItemTransactionMap với transaction mới chứa item này
				e.ItemTransactionMap[item] = append(e.ItemTransactionMap[item], index+1)
			}
		}

		transaction.Items = filteredItems
		transaction.Utilities = filteredUtilities
	}
}

func (e *EMHUN) SortItemsInTransactions() {
	for _, transaction := range e.Transactions {
		itemUtilityMap := make(map[int]float64) // Sửa giá trị map từ int thành float64
		for i, item := range transaction.Items {
			itemUtilityMap[item] = transaction.Utilities[i]
		}

		var positiveItems []int
		var hybridItems []int
		var negativeItems []int

		for _, item := range transaction.Items {
			if e.Rho[item] {
				positiveItems = append(positiveItems, item)
			} else if e.Delta[item] {
				hybridItems = append(hybridItems, item)
			} else if e.Eta[item] {
				negativeItems = append(negativeItems, item)
			}
		}

		positiveItems = e.sortItemsByRTWU(positiveItems)
		hybridItems = e.sortItemsByRTWU(hybridItems)
		negativeItems = e.sortItemsByRTWU(negativeItems)

		sortedItems := append(append(positiveItems, hybridItems...), negativeItems...)

		var sortedUtilities []float64 // Sửa từ int thành float64
		for _, item := range sortedItems {
			sortedUtilities = append(sortedUtilities, itemUtilityMap[item])
		}

		transaction.Items = sortedItems
		transaction.Utilities = sortedUtilities
	}
}
func (e *EMHUN) RebuildItemTransactionMap(oldIndexMap map[*models.Transaction]int) {
	// Xóa map cũ và tạo lại
	e.ItemTransactionMap = make(map[int][]int)

	// Cập nhật lại vị trí transactions chứa từng item
	for newIndex, transaction := range e.Transactions {
		for _, item := range transaction.Items {
			e.ItemTransactionMap[item] = append(e.ItemTransactionMap[item], newIndex+1) // Vị trí mới (bắt đầu từ 1)
		}
	}

	fmt.Println("Updated Item Transaction Map after sorting:")
	e.PrintItemTransactionMap()
}

func (e *EMHUN) SortTransactionsByTWU() {
	fmt.Println("\nSorting transactions by total RLU of items...")

	// Lưu trữ vị trí ban đầu của các transactions
	oldIndexMap := make(map[*models.Transaction]int)
	for i, transaction := range e.Transactions {
		oldIndexMap[transaction] = i + 1 // Lưu index cũ (bắt đầu từ 1)
	}

	// Sắp xếp transactions theo tổng RLU
	sort.Slice(e.Transactions, func(i, j int) bool {
		tuI := utility.CalculateTransactionUtility(e.Transactions[i])
		tuJ := utility.CalculateTransactionUtility(e.Transactions[j])

		// Sắp xếp tăng dần theo tổng RLU
		return tuI < tuJ
	})

	// Cập nhật lại ItemTransactionMap
	e.RebuildItemTransactionMap(oldIndexMap)
}

func (e *EMHUN) sortItemsByRTWU(items []int) []int {
	sort.Slice(items, func(i, j int) bool {
		return e.UtilityArray.GetRTWU(items[i]) < e.UtilityArray.GetRTWU(items[j])
	})
	return items
}

func (e *EMHUN) identifyPrimaryItems() {
	for _, item := range e.SortedSecondary {
		if e.UtilityArray.GetRSU(item) >= e.MinUtility {
			e.PrimaryItems = append(e.PrimaryItems, item)
		}
	}
}

func (e *EMHUN) getTypeOrder(item int) int {
	if e.Rho[item] {
		return 1
	}
	if e.Delta[item] {
		return 2
	}
	if e.Eta[item] {
		return 3
	}
	return int(^uint(0) >> 1)
}

func (e *EMHUN) unionKeys(map1, map2 map[int]bool) map[int]bool {
	unionMap := make(map[int]bool)

	for k := range map1 {
		unionMap[k] = true
	}

	for k := range map2 {
		unionMap[k] = true
	}

	return unionMap
}

func (e *EMHUN) keys(m map[int]bool) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func convertSliceToMap(slice []int) map[int]bool {
	result := make(map[int]bool)
	for _, item := range slice {
		result[item] = true
	}
	return result
}
