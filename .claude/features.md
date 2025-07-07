# Features & Roadmap

## Current Features

### Account Management
- [x] Create bank accounts
- [x] View account details
- [x] Update account information
- [x] Delete accounts
- [x] Track account balances

### Transaction Management
- [x] Record income/expense transactions
- [x] Categorize transactions
- [x] View transaction history
- [x] Update transaction details
- [x] Delete transactions

### Person Management
- [x] Add people for shared expenses
- [x] Manage contact information
- [x] View person profiles

### Bill Management
- [x] Create recurring bills
- [x] Track due dates
- [x] Mark bills as paid
- [x] View upcoming bills

### Credit Card Management
- [x] Link credit cards to accounts
- [x] Track credit card balances
- [x] Monitor credit utilization
- [x] Payment tracking

## Planned Features

### Phase 1: Core Enhancements
- [ ] Shared transaction splitting
- [ ] Settlement calculations
- [ ] Budget tracking
- [ ] Expense categories
- [ ] Monthly/yearly reports

### Phase 2: Advanced Features
- [ ] Data import/export (CSV, JSON)
- [ ] Backup and restore
- [ ] Multi-currency support
- [ ] Investment tracking
- [ ] Goal setting and tracking

### Phase 3: Integrations
- [ ] Bank API integrations
- [ ] Calendar integration for bills
- [ ] Email notifications
- [ ] Mobile app companion
- [ ] Web dashboard

## Feature Specifications

### Shared Transactions
```
User Story: As a user, I want to split expenses with friends
so that I can track who owes money to whom.

Acceptance Criteria:
- Split transactions equally or by custom amounts
- Track payment status for each person
- Calculate settlement amounts
- Generate settlement reports
```

### Budget Tracking
```
User Story: As a user, I want to set budgets for categories
so that I can control my spending.

Acceptance Criteria:
- Set monthly/yearly budgets by category
- Track spending against budgets
- Alert when approaching budget limits
- Visualize budget vs actual spending
```

### Reporting
```
User Story: As a user, I want detailed financial reports
so that I can understand my financial patterns.

Acceptance Criteria:
- Monthly income/expense summaries
- Category-wise spending analysis
- Account balance trends
- Export reports to PDF/CSV
```

## Technical Requirements

### Performance
- Application startup < 2 seconds
- Query response time < 500ms
- Support for 10,000+ transactions
- Efficient pagination for large datasets

### Security
- Secure database connections
- Input validation and sanitization
- No sensitive data in logs
- Secure backup encryption

### Usability
- Intuitive menu navigation
- Clear error messages
- Helpful prompts and defaults
- Consistent UI/UX patterns
