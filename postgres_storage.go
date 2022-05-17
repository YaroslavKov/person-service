package main

import (
	"database/sql"
	"fmt"
	"github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
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

func (s *PostgresStorage) GetAll() ([]*Person, error) {
	var pp []*Person

	err := s.db.Select(&pp, `SELECT * FROM person`)
	if err != nil {
		return nil, err
	}

	if len(pp) == 0 {
		return nil, personNotFoundError
	}

	for _, p := range pp {
		var cc []*Communication
		err = s.db.Select(&cc, `SELECT Value
			FROM Communication
			WHERE PersonId = $1`, p.ID.String())
		if err == nil {
			p.Communications = cc
		}
	}

	return pp, nil
}

func (s *PostgresStorage) Add(p *Person) (*Person, error) {
	_, err := s.db.Exec(`INSERT INTO person (id, name) VALUES ($1, $2)`, p.ID.String(), p.Name)
	if err != nil {
		if pgerr, ok := err.(*pgconn.PgError); ok {
			if pgerr.Code == "23505" {
				return nil, personExistError
			}
		}
		return nil, err
	}

	for _, com := range p.Communications {
		_, err = s.db.Exec(`INSERT INTO communication (value, personid) VALUES ($1, $2)`, com.Value, p.ID.String())
	}

	return s.GetPersonByID(p.ID)
}

func (s *PostgresStorage) GetPersonByID(id uuid.UUID) (*Person, error) {
	p := &Person{}
	err := s.db.Get(p, `SELECT * FROM person WHERE id = $1`, id.String())
	if err == sql.ErrNoRows {
		return nil, personNotFoundError
	} else if err != nil {
		return nil, err
	}

	var cc []*Communication
	err = s.db.Select(&cc, fmt.Sprintf(`SELECT Value
			FROM Communication
			WHERE PersonId = '%s'`, p.ID.String()))
	if err == nil {
		p.Communications = cc
	}

	return p, nil
}

func (s *PostgresStorage) GetPersonsByName(name string) ([]*Person, error) {
	pp := []*Person{}
	err := s.db.Select(&pp, `SELECT * FROM person WHERE Name = $1`, name)
	if err != nil {
		return nil, err
	} else if len(pp) == 0 {
		return nil, personNotFoundError
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

	return pp, nil
}

func (s *PostgresStorage) GetPersonsByCommunication(value string) ([]*Person, error) {
	var pIds []string
	rows, err := s.db.Query(`SELECT DISTINCT PersonId
		FROM Communication
		WHERE Value = $1`, value)

	for rows.Next() {
		var pId uuid.UUID
		err = rows.Scan(&pId)
		if err == nil {
			pIds = append(pIds, pId.String())
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	} else if len(pIds) == 0 {
		return nil, personNotFoundError
	}

	query, args, err := sqlx.In(`SELECT * FROM person WHERE Id IN (?)`, pIds)
	if err != nil {
		return nil, err
	}

	var pp []*Person
	query = s.db.Rebind(query)
	err = s.db.Select(&pp, query, args...)
	if err != nil {
		return nil, err
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

	return pp, nil
}

func (s *PostgresStorage) UpdatePerson(p *Person) (*Person, error) {
	_, err := s.GetPersonByID(p.ID)
	if err != nil {
		return nil, err
	}
	_, err = s.DeletePerson(p.ID)
	if err != nil {
		return nil, err
	}

	_, err = s.Add(p)
	if err == nil {
		return nil, err
	}
	return s.GetPersonByID(p.ID)
}

func (s *PostgresStorage) DeletePerson(id uuid.UUID) (*Person, error) {
	_, err := s.db.Exec(`
		DELETE
		FROM communication
		WHERE personid = $1`, id.String())
	_, err = s.db.Exec(`
		DELETE
		FROM person
		WHERE id = $1`, id.String())

	if err != nil {
		return nil, err
	}
	return s.GetPersonByID(id)
}
