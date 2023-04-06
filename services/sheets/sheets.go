package sheets

import (
	"context"
	"golang.org/x/oauth2/google"
	"log"
	"os"
	"strconv"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type ISheetsAPI interface {
	LoadValues() []Expense
	Write(values [][]interface{}) bool
	LoadValuesByUsername(username string) []Expense
	LoadTotalByUsers(onlyActive bool) []AmountByUser
}
type SheetService struct {
	client *sheets.Service
}

type AmountByUser struct {
	Name  string
	Total float64
}

type Expense struct {
	Id       int64
	Amount   float64
	Reason   string
	From     string
	Date     string
	Username string
	Active   bool
}

func initService() *sheets.Service {
	ctx := context.Background()
	serviceAccount, err := os.ReadFile("./serviceAccount.json")

	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	// Создаем объект конфигурации из JSON-ключа
	config, err := google.JWTConfigFromJSON(serviceAccount, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Failed to get json config %v", err)
	}

	// Создаем клиент API для доступа к Google Sheets API
	client := config.Client(ctx)
	sheetsService, err := sheets.NewService(ctx, option.WithHTTPClient(client))

	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}
	log.Println("Google API Successfully initialize!")

	return sheetsService
}

func NewService() ISheetsAPI {
	return &SheetService{
		client: initService(),
	}
}

func (s *SheetService) Write(values [][]interface{}) bool {
	spreadsheetId := os.Getenv("GOOGLE_SHEET_ID")
	range2 := "1!A2:G"
	// How the input data should be interpreted.
	valueInputOption := "RAW"

	// How the input data should be inserted.
	insertDataOption := "INSERT_ROWS"
	rb := &sheets.ValueRange{
		Values: values,
	}
	_, err := s.client.Spreadsheets.Values.Append(spreadsheetId, range2, rb).ValueInputOption(valueInputOption).InsertDataOption(insertDataOption).Do()

	if err != nil {
		log.Printf("Unable to write data to sheet: %v\n", err)
		return false
	}
	return true
}

func (s *SheetService) LoadValues() []Expense {
	spreadsheetId := os.Getenv("GOOGLE_SHEET_ID")
	readRange := "1!A2:G"
	resp, err := s.client.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Printf("Unable to retrieve data from sheet: %v\n", err)
	}

	expenses := make([]Expense, len(resp.Values))

	for i, row := range resp.Values {
		id, _ := strconv.ParseInt(row[0].(string), 10, 64)
		amount := fixAmount(row[1].(string))
		isActive, _ := strconv.ParseBool(row[6].(string))

		expenses[i] = Expense{
			Id:       id,
			Amount:   amount,
			Reason:   row[2].(string),
			From:     row[3].(string),
			Date:     row[4].(string),
			Username: row[5].(string),
			Active:   isActive,
		}
	}

	return expenses
}

func (s *SheetService) LoadValuesByUsername(username string) []Expense {
	expenses := s.LoadValues()
	var expensesByUsername []Expense

	for _, expense := range expenses {
		if expense.Username == username {
			expensesByUsername = append(expensesByUsername, expense)
		}
	}

	return expensesByUsername
}

func (s *SheetService) LoadTotalByUsers(onlyActive bool) []AmountByUser {
	expensesByUserName := map[string]float64{}

	for _, row := range s.LoadValues() {
		if !onlyActive || row.Active {
			expensesByUserName[row.Username] += row.Amount
		}
	}

	result := make([]AmountByUser, 0, len(expensesByUserName))
	for name, amount := range expensesByUserName {
		result = append(result, AmountByUser{Name: name, Total: amount})
	}

	return result
}

func fixAmount(str string) float64 {
	if str == "" {
		return 0
	}
	amount, err := strconv.ParseFloat(strings.Replace(str, ",", ".", -1), 64)
	if err != nil {
		log.Printf("Failed to parse amount: %v, original value: %s\n", err, str)
	}
	return amount
}
