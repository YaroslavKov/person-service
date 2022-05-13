package main

import (
	"fmt"
	"github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
	"strings"
)

type PostgresStorage struct {
	db *sqlx.DB
}

const connectionString = "postgres://postgres:postgres@localhost:5432/wasfaty?sslmode=disable"

func NewPostgresStorage() (*PostgresStorage, error) {
	db, err := sqlx.Connect("pgx", connectionString)
	if err != nil {
		return nil, err
	}
	return &PostgresStorage{db}, nil
}

func (s *PostgresStorage) GetAll() []*Person {
	var pp []*Person

	err := s.db.Select(&pp, `SELECT * FROM person`)
	if err != nil {
		return nil
	}

	for _, p := range pp {
		var cc []*Communication
		err = s.db.Select(&cc, fmt.Sprintf(`SELECT Value
			FROM Communication
			WHERE PersonId = '%s'`, p.ID.String()))
		if err == nil {
			p.Communications = cc
		}
	}

	return pp
}

func (s *PostgresStorage) Add(p *Person) error {
	_, err := s.db.Exec(fmt.Sprintf(`INSERT INTO person (id, name) VALUES ('%s', '%s')`, p.ID.String(), p.Name))
	if err != nil {
		if pgerr, ok := err.(*pgconn.PgError); ok {
			if pgerr.Code == "23505" {
				return personExistError
			}
		}
		return err
	}

	for _, com := range p.Communications {
		_, err = s.db.Exec(fmt.Sprintf(`INSERT INTO communication (value, personid) VALUES ('%s', '%s')`, com.Value, p.ID.String()))
	}

	return err
}

func (s *PostgresStorage) GetPersonById(id uuid.UUID) *Person {
	var pp []*Person
	err := s.db.Select(&pp, fmt.Sprintf(`SELECT * FROM person WHERE id = '%s' LIMIT 1`, id.String()))
	if err != nil || len(pp) == 0 {
		return nil
	}
	p := pp[0]

	var cc []*Communication
	err = s.db.Select(&cc, fmt.Sprintf(`SELECT Value
			FROM Communication
			WHERE PersonId = '%s'`, p.ID.String()))
	if err == nil {
		p.Communications = cc
	}

	return p
}

func (s *PostgresStorage) GetPersonsByName(name string) []*Person {
	var pp []*Person

	err := s.db.Select(&pp, fmt.Sprintf(`SELECT * FROM person WHERE Name = '%s'`, name))
	if err != nil {
		return nil
	}

	for _, p := range pp {
		var cc []*Communication
		err = s.db.Select(&cc, fmt.Sprintf(`SELECT Value
			FROM Communication
			WHERE PersonId = '%s'`, p.ID.String()))
		if err == nil {
			p.Communications = cc
		}
	}

	return pp
}

func (s *PostgresStorage) GetPersonsByCommunication(value string) []*Person {
	var pIds []string

	rows, err := s.db.Query(fmt.Sprintf(`SELECT DISTINCT PersonId
		FROM Communication
		WHERE Value = '%s'`, value))

	for rows.Next() {
		var pId uuid.UUID
		err = rows.Scan(&pId)
		if err == nil {
			pIds = append(pIds, "'"+pId.String()+"'")
		}
	}
	err = rows.Err()
	if err != nil || len(pIds) == 0 {
		return nil
	}

	var pp []*Person
	err = s.db.Select(&pp, fmt.Sprintf(`SELECT * FROM person WHERE Id IN (%s)`, strings.Join(pIds, ", ")))
	if err != nil {
		return nil
	}

	for _, p := range pp {
		var cc []*Communication
		err = s.db.Select(&cc, fmt.Sprintf(`SELECT Value
			FROM Communication
			WHERE PersonId = '%s'`, p.ID.String()))
		if err == nil {
			p.Communications = cc
		}
	}

	return pp
}

func (s *PostgresStorage) UpdatePerson(p *Person) bool {
	if s.DeletePerson(p.ID) {
		err := s.Add(p)
		if err == nil {
			return true
		}
	}
	return false
}

func (s *PostgresStorage) DeletePerson(id uuid.UUID) bool {
	param := id.String()
	_, err := s.db.Exec(fmt.Sprintf(`
		DELETE
		FROM communication
		WHERE personid = '%s';

		DELETE
		FROM person
		WHERE id = '%s'`, param, param))

	if err != nil {
		return false
	}
	return true
}
