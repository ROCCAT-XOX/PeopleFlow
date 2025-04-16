package repository

import (
	"context"
	"time"

	"PeoplePilot/backend/db"
	"PeoplePilot/backend/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EmployeeRepository enthält alle Datenbankoperationen für das Employee-Modell
type EmployeeRepository struct {
	collection *mongo.Collection
}

// NewEmployeeRepository erstellt ein neues EmployeeRepository
func NewEmployeeRepository() *EmployeeRepository {
	return &EmployeeRepository{
		collection: db.GetCollection("employees"),
	}
}

// Create erstellt einen neuen Mitarbeiter
func (r *EmployeeRepository) Create(employee *model.Employee) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prüfen, ob bereits ein Mitarbeiter mit dieser E-Mail existiert
	count, err := r.collection.CountDocuments(ctx, bson.M{"email": employee.Email})
	if err != nil {
		return err
	}
	if count > 0 {
		return mongo.ErrNoDocuments // Fehlertyp hier nicht optimal, aber einfach zu erkennen
	}

	// Standardwerte setzen für fehlende Zeitstempel
	if employee.CreatedAt.IsZero() {
		employee.CreatedAt = time.Now()
	}
	if employee.UpdatedAt.IsZero() {
		employee.UpdatedAt = time.Now()
	}

	result, err := r.collection.InsertOne(ctx, employee)
	if err != nil {
		return err
	}

	employee.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByID findet einen Mitarbeiter anhand seiner ID
func (r *EmployeeRepository) FindByID(id string) (*model.Employee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var employee model.Employee
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&employee)
	if err != nil {
		return nil, err
	}

	return &employee, nil
}

// FindByEmail findet einen Mitarbeiter anhand seiner E-Mail
func (r *EmployeeRepository) FindByEmail(email string) (*model.Employee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var employee model.Employee
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&employee)
	if err != nil {
		return nil, err
	}

	return &employee, nil
}

// FindAll findet alle Mitarbeiter
func (r *EmployeeRepository) FindAll() ([]*model.Employee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Optionen für die Sortierung nach Nachname und Vorname
	opts := options.Find().SetSort(bson.D{
		{Key: "lastName", Value: 1},
		{Key: "firstName", Value: 1},
	})

	var employees []*model.Employee
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var employee model.Employee
		if err := cursor.Decode(&employee); err != nil {
			return nil, err
		}
		employees = append(employees, &employee)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return employees, nil
}

// FindManagers findet alle Mitarbeiter, die als Manager fungieren können
func (r *EmployeeRepository) FindManagers() ([]*model.Employee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Hier könnten eigentlich noch weitere Filter eingesetzt werden,
	// z.B. nur Mitarbeiter ab einer bestimmten Position als Manager zulassen
	var managers []*model.Employee
	cursor, err := r.collection.Find(ctx, bson.M{"status": model.EmployeeStatusActive})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var employee model.Employee
		if err := cursor.Decode(&employee); err != nil {
			return nil, err
		}
		managers = append(managers, &employee)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return managers, nil
}

// Update aktualisiert einen Mitarbeiter
func (r *EmployeeRepository) Update(employee *model.Employee) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// UpdatedAt-Zeitstempel aktualisieren
	employee.UpdatedAt = time.Now()

	// Prüfen, ob bereits ein anderer Mitarbeiter mit dieser E-Mail existiert
	var existingEmployee model.Employee
	err := r.collection.FindOne(ctx, bson.M{
		"email": employee.Email,
		"_id":   bson.M{"$ne": employee.ID},
	}).Decode(&existingEmployee)

	// Wenn ein Dokument gefunden wurde, bedeutet das, dass die E-Mail bereits verwendet wird
	if err == nil {
		return mongo.ErrNoDocuments
	}
	// Wenn der Fehler nicht "nicht gefunden" ist, ist es ein anderer Fehler
	if err != mongo.ErrNoDocuments {
		return err
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": employee.ID},
		bson.M{"$set": employee},
	)
	return err
}

// Delete löscht einen Mitarbeiter
func (r *EmployeeRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

// FindByDepartment findet alle Mitarbeiter einer bestimmten Abteilung
func (r *EmployeeRepository) FindByDepartment(department model.Department) ([]*model.Employee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var employees []*model.Employee
	cursor, err := r.collection.Find(ctx, bson.M{"department": department})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var employee model.Employee
		if err := cursor.Decode(&employee); err != nil {
			return nil, err
		}
		employees = append(employees, &employee)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return employees, nil
}

// FindByManager findet alle Mitarbeiter eines bestimmten Managers
func (r *EmployeeRepository) FindByManager(managerID string) ([]*model.Employee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(managerID)
	if err != nil {
		return nil, err
	}

	var employees []*model.Employee
	cursor, err := r.collection.Find(ctx, bson.M{"managerId": objID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var employee model.Employee
		if err := cursor.Decode(&employee); err != nil {
			return nil, err
		}
		employees = append(employees, &employee)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return employees, nil
}

// CountByDepartment zählt die Mitarbeiter pro Abteilung
func (r *EmployeeRepository) CountByDepartment() (map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aggregation für die Zählung nach Abteilungen
	pipeline := []bson.M{
		{"$group": bson.M{
			"_id":   "$department",
			"count": bson.M{"$sum": 1},
		}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Ergebnisse verarbeiten
	result := make(map[string]int)
	for cursor.Next(ctx) {
		var item struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		result[item.ID] = item.Count
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
