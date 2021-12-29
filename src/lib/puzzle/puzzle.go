package lib_puzzle

import (
	"fmt"
	"strconv"
)

//MUST check if len(text) == 81 before any further using
func IsBoardSolvable(boardText string) bool {
	board := parseText(boardText)
	if !hasEmptyCell(board) || len(boardText) != 81 || boardText == "000000000000000000000000000000000000000000000000000000000000000000000000000000000" {
		return false
	}
	return solveBoard(board)
}

func solveBoard(board *[9][9]int) bool {
	if !hasEmptyCell(board) {
		return true
	}
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if board[i][j] == 0 {
				for candidate := 9; candidate >= 1; candidate-- {
					board[i][j] = candidate
					if isBoardValid(board) {
						if solveBoard(board) {
							return true
						}
						board[i][j] = 0
					} else {
						board[i][j] = 0
					}
				}
				return false
			}
		}
	}
	return false
}

func hasEmptyCell(board *[9][9]int) bool {
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if board[i][j] == 0 {
				return true
			}
		}
	}
	return false
}

func isBoardValid(board *[9][9]int) bool {

	//check duplicates by row
	for row := 0; row < 9; row++ {
		counter := [10]int{}
		for col := 0; col < 9; col++ {
			counter[board[row][col]]++
		}
		if hasDuplicate(counter) {
			return false
		}
	}

	//check duplicates by column
	for row := 0; row < 9; row++ {
		counter := [10]int{}
		for col := 0; col < 9; col++ {
			counter[board[col][row]]++
		}
		if hasDuplicate(counter) {
			return false
		}
	}

	//check 3x3 section
	for i := 0; i < 9; i += 3 {
		for j := 0; j < 9; j += 3 {
			counter := [10]int{}
			for row := i; row < i+3; row++ {
				for col := j; col < j+3; col++ {
					counter[board[row][col]]++
				}
				if hasDuplicate(counter) {
					return false
				}
			}
		}
	}

	return true
}

func hasDuplicate(counter [10]int) bool {
	for i, count := range counter {
		if i == 0 {
			continue
		}
		if count > 1 {
			return true
		}
	}
	return false
}

func parseText(text string) *[9][9]int {
	var result [9][9]int
	i := 0
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			string := string(text[i])
			int, _ := strconv.Atoi(string)
			result[row][col] = int
			i++
		}
	}
	return &result
}

func PrintBoard(board [9][9]int) {
	fmt.Println("+-------+-------+-------+")
	for row := 0; row < 9; row++ {
		fmt.Print("| ")
		for col := 0; col < 9; col++ {
			if col == 3 || col == 6 {
				fmt.Print("| ")
			}
			if board[row][col] == 0 {
				fmt.Printf("* ")
			} else {
				fmt.Printf("%d ", board[row][col])
			}
			if col == 8 {
				fmt.Print("|")
			}
		}
		if row == 2 || row == 5 || row == 8 {
			fmt.Println("\n+-------+-------+-------+")
		} else {
			fmt.Println()
		}
	}
}
