package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var mu sync.Mutex

const (
	limitPerRequest = 5_000
)

func GenerateUsers(amount int) ([]User, error) {
	ug := usersGenerator{
		startTime: time.Now(),
		usersId: make(map[string]bool, amount),
		users: make([]User, 0),
		amount: amount,
		err: nil,
		wg: &sync.WaitGroup{},
		timeLimit: time.Second * 2,
	}
	ug.generate()
	return ug.users, ug.err
}

func getUsers(amount int) ([]UserApi, error) {
	url := "https://randomuser.me/api/?format=json&inc=gender,name,email,login&results=" + strconv.Itoa(amount)
	res, err := http.Get(url)
	var resStruct ResponseRandomUsersApi
	if err != nil {
		return resStruct.Results, err
	}
	if res.StatusCode != http.StatusOK {
		return resStruct.Results, fmt.Errorf(res.Status)
	}
	if resStruct.Error != "" {
		return resStruct.Results, fmt.Errorf(resStruct.Error)
	}
	err = json.NewDecoder(res.Body).Decode(&resStruct)
	if err != nil {
		return resStruct.Results, err
	}
	return resStruct.Results, nil
}

type usersGenerator struct {
	startTime time.Time
	usersId map[string]bool
	users []User
	amount int
	amountAdded int
	err error
	wg *sync.WaitGroup
	timeLimit time.Duration
}

func (ug *usersGenerator) generate() {
	requestsMade := 0
	for i := 0; i < ug.amount; i += limitPerRequest {
		requestsMade++
		if ug.err != nil {
			return
		}
		ug.wg.Add(1)
		if (ug.amount - ug.amountAdded) > limitPerRequest {
			go ug.generateBatch(limitPerRequest)
			continue
		}
		go ug.generateBatch(ug.amount - ug.amountAdded)
	}
	ug.wg.Wait()
}

func (ug *usersGenerator) generateBatch(amount int)  {
	defer ug.wg.Done()
	usersApi, err := getUsers(amount)
	if err != nil {
		ug.checkRetry(amount, err)
		return
	}
	addedInBatch := 0
	var wg sync.WaitGroup
	arrayLen := len(usersApi)
	batchSize := 1_000
	if batchSize > arrayLen {
			batchSize = arrayLen
	}
	for i := 0; i < arrayLen; i += batchSize {
			hi := i + batchSize
			if hi > arrayLen {
					hi = arrayLen
			}
			wg.Add(1)
			go ug.formatAddUsersBatch(i, hi, usersApi, &addedInBatch, &wg)
	}
	wg.Wait()
	println("added in batch", addedInBatch)
	if (addedInBatch < amount) {
		ug.checkRetry(amount - addedInBatch, fmt.Errorf("not enough users"))
	}
}

func (ug *usersGenerator) formatAddUsersBatch(lo int, hi int, usersApi []UserApi, amountAdded *int, wg *sync.WaitGroup) {
	defer wg.Done()
	added := 0
	for i := lo; i < hi; i++ {
		added += ug.formatAddUser(usersApi[i])
	}
	println("added in mini batch", added)
	mu.Lock()
	*amountAdded += added
	mu.Unlock()
}

func (ug *usersGenerator) formatAddUser(userApi UserApi) int {
	user := userApi.convertToRawUser()
	mu.Lock()
	if _, ok := ug.usersId[user.UUID]; !ok {
		ug.usersId[user.UUID] = true
		ug.users = append(ug.users, user)
		mu.Unlock()
		return 1
	}
	mu.Unlock()
	return 0
}

func (ug *usersGenerator) checkRetry(amount int, err error) {
	println(err.Error())
	if time.Since(ug.startTime) > ug.timeLimit {
		if ug.err != nil {
			return
		}
		mu.Lock()
		ug.err = err
		mu.Unlock()
		return
	}
	ug.wg.Add(1)
	time.Sleep(100 * time.Millisecond)
	go ug.generateBatch(amount)
}

