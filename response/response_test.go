package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestConstants(t *testing.T) {
	assert.Equal(t, 0, CodeSuccess, "CodeSuccess should be 0")
	assert.Equal(t, -1, CodeSystemError, "CodeSystemError should be -1")
}

func TestResponse_Structure(t *testing.T) {
	// 测试Response结构体
	data := "test data"
	resp := Response[string]{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	}

	assert.Equal(t, CodeSuccess, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, data, resp.Data)
}

func TestSuccess(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		Success(c, "test data", "operation successful")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response[string]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSuccess, response.Code)
	assert.Equal(t, "operation successful", response.Message)
	assert.Equal(t, "test data", response.Data)
}

func TestSuccess_DefaultMessage(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		Success(c, "test data")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response[string]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSuccess, response.Code)
	assert.Equal(t, "success", response.Message)
	assert.Equal(t, "test data", response.Data)
}

func TestCreated(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		Created(c, map[string]string{"id": "123"}, "resource created")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response Response[map[string]string]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSuccess, response.Code)
	assert.Equal(t, "resource created", response.Message)
	assert.Equal(t, "123", response.Data["id"])
}

func TestCreated_DefaultMessage(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		Created(c, "created data")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response Response[string]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSuccess, response.Code)
	assert.Equal(t, "created", response.Message)
	assert.Equal(t, "created data", response.Data)
}

func TestNoContent(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		NoContent(c, "no content available")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestNoContent_DefaultMessage(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		NoContent(c)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestError_BusinessError(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		testErr := &testError{msg: "business logic error"}
		Error(c, testErr, 1001, "error details")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response[string]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1001, response.Code)
	assert.Equal(t, "business logic error", response.Message)
	assert.Equal(t, "error details", response.Data)
}

func TestError_SystemError(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		testErr := &testError{msg: "system failure"}
		Error(c, testErr, 0, "system error details")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response Response[string]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSystemError, response.Code)
	assert.Equal(t, "system failure", response.Message)
	assert.Equal(t, "system error details", response.Data)
}

func TestError_NilError(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		Error(c, nil, 1001, "fallback data")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response[string]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSuccess, response.Code)
	assert.Equal(t, "success", response.Message)
	assert.Equal(t, "fallback data", response.Data)
}

func TestErrorNoData(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		testErr := &testError{msg: "error without data"}
		ErrorNoData(c, testErr, 400)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 400, response.Code)
	assert.Equal(t, "error without data", response.Message)
	assert.Nil(t, response.Data)
}

func TestNotFound(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		NotFound(c, "user not found")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Response[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 404, response.Code)
	assert.Equal(t, "user not found", response.Message)
	assert.Nil(t, response.Data)
}

func TestNotFound_DefaultMessage(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		NotFound(c)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Response[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 404, response.Code)
	assert.Equal(t, "resource not found", response.Message)
	assert.Nil(t, response.Data)
}

func TestForbidden(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		Forbidden(c, "access denied")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response Response[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 403, response.Code)
	assert.Equal(t, "access denied", response.Message)
	assert.Nil(t, response.Data)
}

func TestForbidden_DefaultMessage(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		Forbidden(c)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response Response[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 403, response.Code)
	assert.Equal(t, "forbidden", response.Message)
	assert.Nil(t, response.Data)
}

func TestBadRequest(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		BadRequest(c, "invalid parameters")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response Response[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 400, response.Code)
	assert.Equal(t, "invalid parameters", response.Message)
	assert.Nil(t, response.Data)
}

func TestBadRequest_DefaultMessage(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		BadRequest(c)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response Response[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 400, response.Code)
	assert.Equal(t, "bad request", response.Message)
	assert.Nil(t, response.Data)
}

func TestUnprocessableEntity(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		UnprocessableEntity(c, "validation failed")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var response Response[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 422, response.Code)
	assert.Equal(t, "validation failed", response.Message)
	assert.Nil(t, response.Data)
}

func TestUnprocessableEntity_DefaultMessage(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		UnprocessableEntity(c)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var response Response[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 422, response.Code)
	assert.Equal(t, "unprocessable entity", response.Message)
	assert.Nil(t, response.Data)
}

func TestSystemError(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		testErr := &testError{msg: "database connection failed"}
		SystemError(c, testErr, "error context")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response Response[string]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSystemError, response.Code)
	assert.Equal(t, "database connection failed", response.Message)
	assert.Equal(t, "error context", response.Data)
}

func TestSystemError_NilError(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		SystemError(c, nil, "fallback data")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response Response[string]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSystemError, response.Code)
	assert.Equal(t, "system error", response.Message)
	assert.Equal(t, "fallback data", response.Data)
}

func TestSystemErrorNoData(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		testErr := &testError{msg: "internal server error"}
		SystemErrorNoData(c, testErr)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response Response[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSystemError, response.Code)
	assert.Equal(t, "internal server error", response.Message)
	assert.Nil(t, response.Data)
}

func TestSystemErrorNoData_NilError(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		SystemErrorNoData(c, nil)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response Response[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSystemError, response.Code)
	assert.Equal(t, "system error", response.Message)
	assert.Nil(t, response.Data)
}

func TestResponse_WithStructData(t *testing.T) {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		user := User{ID: 1, Name: "John"}
		Success(c, user, "user retrieved")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response[User]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSuccess, response.Code)
	assert.Equal(t, "user retrieved", response.Message)
	assert.Equal(t, 1, response.Data.ID)
	assert.Equal(t, "John", response.Data.Name)
}

func TestResponse_JSONSerialization(t *testing.T) {
	type TestData struct {
		Value string `json:"value"`
	}

	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		data := TestData{Value: "test"}
		Success(c, data, "json test")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 验证JSON结构
	jsonStr := w.Body.String()
	assert.Contains(t, jsonStr, `"code":0`)
	assert.Contains(t, jsonStr, `"message":"json test"`)
	assert.Contains(t, jsonStr, `"data":{"value":"test"}`)

	// 验证可以正确解析
	var response Response[TestData]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSuccess, response.Code)
	assert.Equal(t, "json test", response.Message)
	assert.Equal(t, "test", response.Data.Value)
}

// 辅助测试类型
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
