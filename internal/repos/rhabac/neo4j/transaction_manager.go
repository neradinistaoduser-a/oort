package neo4j

import (
	"context"
	"log"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"go.opentelemetry.io/otel"
)

type TransactionManager struct {
	driver neo4j.Driver
	dbName string
}

func NewTransactionManager(uri, dbName string) (*TransactionManager, error) {
	driver, err := neo4j.NewDriver(uri, neo4j.NoAuth())
	if err != nil {
		return nil, err
	}
	return &TransactionManager{
		driver: driver,
		dbName: dbName,
	}, nil
}

type TransactionFunction func(transaction neo4j.Transaction) (interface{}, error)

func (manager *TransactionManager) WriteTransaction(ctx context.Context, cypher string, params map[string]interface{}) error {
	tracer := otel.Tracer("oort.neo4j")
	ctx, span := tracer.Start(ctx, "neo4j.WriteTransaction")
	defer span.End()

	_, err := manager.writeTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(cypher, params)
		if err != nil {
			_ = transaction.Rollback()
		}
		if result == nil {
			return nil, nil
		}
		return nil, result.Err()
	})
	return err
}

func (manager *TransactionManager) WriteTransactions(ctx context.Context, cyphers []string, params []map[string]interface{}) error {
	tracer := otel.Tracer("oort.neo4j")
	ctx, span := tracer.Start(ctx, "neo4j.WriteTransaction")
	defer span.End()

	_, err := manager.writeTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		var txErr error = nil
		for i := range cyphers {
			cypher := cyphers[i]
			param := params[i]
			result, err := transaction.Run(cypher, param)
			if err != nil || result.Err() != nil {
				_ = transaction.Rollback()
				if err != nil {
					txErr = err
				} else {
					txErr = result.Err()
				}
				break
			}
		}
		return txErr, nil
	})
	return err
}

func (manager *TransactionManager) ReadTransaction(ctx context.Context, cypher string, params map[string]interface{}) (interface{}, error) {
	tracer := otel.Tracer("oort.neo4j")
	ctx, span := tracer.Start(ctx, "neo4j.ReadTransaction")
	defer span.End()

	return manager.readTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(cypher, params)
		if err != nil {
			return nil, err
		}
		if result.Err() != nil {
			return nil, result.Err()
		}
		return result.Collect()
	})
}

func (manager *TransactionManager) writeTransaction(txFunc TransactionFunction) (interface{}, error) {
	session := manager.driver.NewSession(neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: manager.dbName})
	defer func(session neo4j.Session) {
		err := session.Close()
		if err != nil {
			log.Println(err)
		}
	}(session)

	result, err := session.WriteTransaction(neo4j.TransactionWork(txFunc))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (manager *TransactionManager) readTransaction(txFunc TransactionFunction) (interface{}, error) {
	session := manager.driver.NewSession(neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: manager.dbName})
	defer func(session neo4j.Session) {
		err := session.Close()
		if err != nil {
			log.Println(err)
		}
	}(session)

	result, err := session.ReadTransaction(neo4j.TransactionWork(txFunc))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (manager *TransactionManager) Stop() {
	err := manager.driver.Close()
	if err != nil {
		log.Println("error while closing neo4j conn: ", err)
	}
}
