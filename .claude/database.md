# Database Setup & Schema

## MongoDB Configuration

### Environment Variables
```bash
MONGODB_URI=mongodb://admin:password@localhost:27017/financli?authSource=admin
MONGODB_DATABASE=financli
```

### Docker Setup
```bash
docker-compose up -d  # Start MongoDB
docker-compose logs   # Check logs
docker-compose down   # Stop MongoDB
```

## Collections Schema

### accounts
```json
{
  "_id": ObjectId,
  "name": "Personal Checking",
  "bank_name": "Chase Bank",
  "account_number": "****1234",
  "balance": 5000.00,
  "created_at": ISODate(),
  "updated_at": ISODate()
}
```

### transactions
```json
{
  "_id": ObjectId,
  "amount": 150.00,
  "description": "Groceries",
  "date": ISODate(),
  "account_id": ObjectId,
  "type": "expense",
  "category": "food",
  "created_at": ISODate(),
  "updated_at": ISODate()
}
```

### people
```json
{
  "_id": ObjectId,
  "name": "John Doe",
  "email": "john@example.com",
  "created_at": ISODate(),
  "updated_at": ISODate()
}
```

### bills
```json
{
  "_id": ObjectId,
  "name": "Netflix Subscription",
  "amount": 15.99,
  "due_date": ISODate(),
  "frequency": "monthly",
  "account_id": ObjectId,
  "is_active": true,
  "created_at": ISODate(),
  "updated_at": ISODate()
}
```

### credit_cards
```json
{
  "_id": ObjectId,
  "name": "Chase Sapphire",
  "last_four": "5678",
  "credit_limit": 10000.00,
  "current_balance": 2500.00,
  "account_id": ObjectId,
  "created_at": ISODate(),
  "updated_at": ISODate()
}
```

## Indexes
```javascript
// Performance indexes
db.transactions.createIndex({ "date": -1 })
db.transactions.createIndex({ "account_id": 1 })
db.transactions.createIndex({ "type": 1 })
db.bills.createIndex({ "due_date": 1 })
db.accounts.createIndex({ "name": 1 })
```

## Connection Management
- Use connection pooling
- Implement proper timeout handling
- Add retry logic for transient failures
- Use MongoDB transactions for multi-document operations
