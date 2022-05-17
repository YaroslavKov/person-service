package main

import (
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestAddPerson(t *testing.T) {
	server := NewServer(NewInMemoryPersonStorage())

	t.Run("wrong content type", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/person", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("wrong person model", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/person", strings.NewReader(
			`<person>
				<id>02a883a3-13c4-4624-bbba-edc744f69534</id>
				<name>Joe</name>
				<communications>
					<communication>
						<value>box@mail.ua</value>
					</communication>
					<communication>
						<value>+380973224562</value>
					</communication>
				</communications>
			</person>`))
		req.Header.Add("Content-Type", contentTypeJSON)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("add person", func(t *testing.T) {
		reqBody :=
			`{
				"id": "02a883a3-13c4-4624-bbba-edc744f69534",
				"name": "Joe",
				"communications": [
					{
						"value": "box@mail.ua"
					},
					{
						"value": "+380974583947"
					}
				]
			}`
		req, _ := http.NewRequest("POST", "/person", strings.NewReader(reqBody))
		req.Header.Add("Content-Type", contentTypeJSON)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusCreated)
		var want, got *Person
		err1 := json.Unmarshal([]byte(reqBody), want)
		err2 := json.Unmarshal(response.Body.Bytes(), got)
		if err1 == nil && err2 == nil && !reflect.DeepEqual(want, got) {
			personsNotEqualError(t, want, got)
		}
	})

	t.Run("add copy", func(t *testing.T) {
		reqBody :=
			`{
			"id": "02a883a3-13c4-4624-bbba-edc744f69534",
			"name": "Joe",
			"communications": [
				{
					"value": "box@mail.ua"
				},
				{
					"value": "+380974583947"
				}
			]
		}`
		req, _ := http.NewRequest("POST", "/person", strings.NewReader(reqBody))
		req.Header.Add("Content-Type", contentTypeJSON)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusUnprocessableEntity)
	})
}

func TestGetPersons(t *testing.T) {
	p1Id := uuid.FromStringOrNil("02a883a3-13c4-4624-bbba-edc744f69534")
	p2Id := uuid.FromStringOrNil("02a883a3-13c4-4624-bbba-edc744f69535")
	p1 := &Person{
		ID:             p1Id,
		Name:           "Joe",
		Communications: []*Communication{{Value: "box@mail.ua"}, {Value: "+380974583947"}},
	}
	p2 := &Person{
		ID:             p2Id,
		Name:           "Joe",
		Communications: []*Communication{{Value: "box@mail.ua"}, {Value: "+380975345865"}},
	}
	data := map[uuid.UUID]*Person{
		p1Id: p1,
		p2Id: p2,
	}
	s := &InMemoryPersonStorage{data}
	server := NewServer(s)

	t.Run("get persons without param", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/person", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("get person by invalid id", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/person?id=123%s", p1Id.String()), nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("get person if id not in storage", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/person?id=02a883a3-13c4-4624-bbba-edc744f69530", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("get person by id", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/person?id=%v", p1Id.String()), nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusOK)
		var want, got *Person = p1, nil
		err := json.Unmarshal(response.Body.Bytes(), got)
		if err == nil && !reflect.DeepEqual(want, got) {
			personsNotEqualError(t, want, got)
		}
	})

	t.Run("get person if name not in storage", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/person?name=Louis", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("get person by name", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/person?name=Joe", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusOK)
		assertPersonArrays(t, response.Body.Bytes(), data)
	})

	t.Run("get person if communication not in storage", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/person?communication=random_string", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("get person by communication", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/person?communication=box@mail.ua", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusOK)
		assertPersonArrays(t, response.Body.Bytes(), data)
	})
}

func TestPutPerson(t *testing.T) {
	server := NewServer(NewInMemoryPersonStorage())

	t.Run("wrong content type", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/person", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("wrong person model", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/person", strings.NewReader(
			`<person>
				<id>02a883a3-13c4-4624-bbba-edc744f69534</id>
				<name>Joe</name>
				<communications>
					<communication>
						<value>box@mail.ua</value>
					</communication>
					<communication>
						<value>+380973224562</value>
					</communication>
				</communications>
			</person>`))
		req.Header.Add("Content-Type", contentTypeJSON)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("put new person", func(t *testing.T) {
		reqBody :=
			`{
				"id": "02a883a3-13c4-4624-bbba-edc744f69534",
				"name": "Joe",
				"communications": [
					{
						"value": "box@mail.ua"
					},
					{
						"value": "+380974583947"
					}
				]
			}`
		req, _ := http.NewRequest("PUT", "/person", strings.NewReader(reqBody))
		req.Header.Add("Content-Type", contentTypeJSON)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusOK)
		var want, got *Person
		err1 := json.NewDecoder(strings.NewReader(reqBody)).Decode(want)
		err2 := json.Unmarshal(response.Body.Bytes(), got)
		if err1 == nil && err2 == nil && !reflect.DeepEqual(want, got) {
			personsNotEqualError(t, want, got)
		}
	})

	t.Run("put exist person", func(t *testing.T) {
		reqBody :=
			`{
				"id": "02a883a3-13c4-4624-bbba-edc744f69534",
				"name": "Louis",
				"communications": [
					{
						"value": "box@mail.ua"
					},
					{
						"value": "+380974583947"
					}
				]
			}`
		req, _ := http.NewRequest("PUT", "/person", strings.NewReader(reqBody))
		req.Header.Add("Content-Type", contentTypeJSON)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusOK)
		var want, got *Person
		err1 := json.NewDecoder(strings.NewReader(reqBody)).Decode(want)
		err2 := json.Unmarshal(response.Body.Bytes(), got)
		if err1 == nil && err2 == nil && !reflect.DeepEqual(want, got) {
			personsNotEqualError(t, want, got)
		}
	})
}

func TestDelete(t *testing.T) {
	pId := uuid.FromStringOrNil("02a883a3-13c4-4624-bbba-edc744f69534")
	p := &Person{
		ID:             pId,
		Name:           "Joe",
		Communications: []*Communication{{Value: "box@mail.ua"}, {Value: "+380974583947"}},
	}
	data := map[uuid.UUID]*Person{pId: p}
	server := NewServer(&InMemoryPersonStorage{data})

	t.Run("wrong id", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/person/123%v", pId.String()), nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("delete Joe", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/person/02a883a3-13c4-4624-bbba-edc744f69530", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("delete Joe", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/person/%v", pId.String()), nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusNoContent)
	})
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}

func personsNotEqualError(t *testing.T, want, got *Person) {
	spew.Config.DisablePointerAddresses = true
	spew.Config.DisableCapacities = true
	t.Errorf("did not get valid response, want \n%s got \n%s", spew.Sdump(want), spew.Sdump(got))
}

func getPersonById(persons []*Person, id uuid.UUID) *Person {
	for _, val := range persons {
		if val.ID == id {
			return val
		}
	}
	return nil
}

func assertPersonArrays(t *testing.T, respBody []byte, data map[uuid.UUID]*Person) {
	t.Helper()

	var want, got []*Person
	json.Unmarshal(respBody, &got)

	for id, val := range data {
		want = append(want, val)
		gotPerson := getPersonById(got, id)
		if gotPerson == nil || !reflect.DeepEqual(val, getPersonById(got, id)) {
			personsNotEqualError(t, val, gotPerson)
		}
	}
}
