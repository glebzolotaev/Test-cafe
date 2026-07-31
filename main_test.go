package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	city := "moscow"
	totalCount := len(cafeList[city])

	requests := []struct {
		count int
		want  int
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: totalCount},
	}

	for _, tc := range requests {
		t.Run(fmt.Sprintf("count=%d", tc.count), func(t *testing.T) {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/cafe?city="+city+"&count="+strconv.Itoa(tc.count), nil)

			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)

			body := strings.TrimSpace(response.Body.String())

			var cafes []string
			if body != "" {
				cafes = strings.Split(body, ",")
			}

			actualCount := len(cafes)
			if body == "" || (len(cafes) == 1 && cafes[0] == "") {
				actualCount = 0
			}

			assert.Equal(t, tc.want, actualCount)
		})
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	city := "moscow"

	requests := []struct {
		search    string
		wantCount int
	}{
		{search: "фасоль", wantCount: 0},
		{search: "кофе", wantCount: 2},
		{search: "вилка", wantCount: 1},
	}

	for _, tc := range requests {
		t.Run(fmt.Sprintf("search=%s", tc.search), func(t *testing.T) {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/cafe?city="+city+"&search="+tc.search, nil)

			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)

			body := strings.TrimSpace(response.Body.String())

			var cafes []string
			if body != "" {
				cafes = strings.Split(body, ",")
			}

			actualCount := len(cafes)
			if body == "" || (len(cafes) == 1 && cafes[0] == "") {
				actualCount = 0
			}

			assert.Equal(t, tc.wantCount, actualCount)

			searchLower := strings.ToLower(tc.search)
			for _, cafeName := range cafes {
				assert.True(t, strings.Contains(strings.ToLower(cafeName), searchLower))
			}
		})
	}
}
