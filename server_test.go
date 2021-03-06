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

const logBody = false

func TestAddPerson(t *testing.T) {
	server := NewServer(NewInMemoryPersonStorage(), logBody)

	t.Run("wrong content type", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/person", nil)
		setRequestAuth(req)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusUnsupportedMediaType)
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
		req.SetBasicAuth(authLogin, authPassword)
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
		req.SetBasicAuth(authLogin, authPassword)
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
		req.SetBasicAuth(authLogin, authPassword)
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
	server := NewServer(s, logBody)

	t.Run("get persons without param", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/person", nil)
		req.SetBasicAuth(authLogin, authPassword)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusOK)
		assertPersonsResponse(t, response.Body.Bytes(), data)
	})

	t.Run("get person by invalid id", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/person/123%s", p1Id.String()), nil)
		req.SetBasicAuth(authLogin, authPassword)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("get person if id not in storage", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/person/02a883a3-13c4-4624-bbba-edc744f69530", nil)
		req.SetBasicAuth(authLogin, authPassword)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("get person by id", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/person/%v", p1Id.String()), nil)
		req.SetBasicAuth(authLogin, authPassword)
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
		req.SetBasicAuth(authLogin, authPassword)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("get person by name", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/person?name=Joe", nil)
		req.SetBasicAuth(authLogin, authPassword)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusOK)
		assertPersonsResponse(t, response.Body.Bytes(), data)
	})

	t.Run("get person if communication not in storage", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/person?communication=random_string", nil)
		req.SetBasicAuth(authLogin, authPassword)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("get person by communication", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/person?communication=box@mail.ua", nil)
		req.SetBasicAuth(authLogin, authPassword)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusOK)
		assertPersonsResponse(t, response.Body.Bytes(), data)
	})
}

func TestPutPerson(t *testing.T) {
	server := NewServer(NewInMemoryPersonStorage(), logBody)

	t.Run("wrong content type", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/person", nil)
		req.SetBasicAuth(authLogin, authPassword)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusUnsupportedMediaType)
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
		req.SetBasicAuth(authLogin, authPassword)
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
		req.SetBasicAuth(authLogin, authPassword)
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
		req.SetBasicAuth(authLogin, authPassword)
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
	server := NewServer(&InMemoryPersonStorage{data}, logBody)

	t.Run("wrong id", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/person/123%v", pId.String()), nil)
		req.SetBasicAuth(authLogin, authPassword)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("delete Joe", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/person/02a883a3-13c4-4624-bbba-edc744f69530", nil)
		req.SetBasicAuth(authLogin, authPassword)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("delete Joe", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/person/%v", pId.String()), nil)
		req.SetBasicAuth(authLogin, authPassword)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatus(t, response.Code, http.StatusNoContent)
	})
}

func TestServerMethods(t *testing.T) {
	p1 := &Person{
		ID:             uuid.FromStringOrNil("02a883a3-13c4-4624-bbba-edc744f69530"),
		Name:           "Joe",
		Communications: []*Communication{{Value: "box@mail.ua"}, {Value: "+380974583947"}},
	}
	p2 := &Person{
		ID:             uuid.FromStringOrNil("02a883a3-13c4-4624-bbba-edc744f69531"),
		Name:           "Joe",
		Communications: []*Communication{{Value: "box@mail.ua"}, {Value: "+380974583947"}},
	}
	p3 := &Person{
		ID:             uuid.FromStringOrNil("02a883a3-13c4-4624-bbba-edc744f69530"),
		Name:           "Joe",
		Communications: []*Communication{{Value: "box@mail.ua"}, {Value: "+380974583947"}},
	}
	p4 := &Person{
		ID:             uuid.FromStringOrNil("02a883a3-13c4-4624-bbba-edc744f69534"),
		Name:           "Joe",
		Communications: []*Communication{{Value: "box@mail.ua"}, {Value: "+380974583947"}},
	}

	t.Run("except 2 arrays of persons", func(t *testing.T) {
		pp1 := []*Person{p1, p2}
		pp2 := []*Person{p3, p4}
		pp3 := exceptPersons(pp1, pp2)
		assertPersonsArrays(t, []*Person{p1}, pp3)
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

func assertPersonsResponse(t *testing.T, respBody []byte, data map[uuid.UUID]*Person) {
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

func assertPersonsArrays(t *testing.T, want, got []*Person) {
	t.Helper()

	for _, val1 := range want {
		catch := false
		for _, val2 := range got {
			if reflect.DeepEqual(val1, val2) {
				catch = true
			}
		}
		if !catch {
			spew.Config.DisablePointerAddresses = true
			spew.Config.DisableCapacities = true
			t.Errorf("did not get valid response, want \n%s", spew.Sdump(val1))
			break
		}
	}
}

func setRequestAuth(r *http.Request) {
	r.SetBasicAuth(authLogin, authPassword)
}
