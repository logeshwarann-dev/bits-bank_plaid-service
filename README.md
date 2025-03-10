# 2️⃣ Plaid Service (Bank Account Linking & Data Fetching)

# 📌 Purpose: Integrates with Plaid to allow users to connect their bank accounts and fetch financial data.

🔹 Tech Stack:

Backend: Golang

Database: PostgreSQL (Stores user-linked bank accounts, transactions)

🔹 Key Responsibilities:

✅ Connect user bank accounts via Plaid Link

✅ Fetch user account details, balances, and transactions

✅ Store Plaid access tokens securely for re-authentication

✅ Webhooks to handle real-time bank account updates

🔹 Endpoints Example:

POST `/plaid/link` → Generate Plaid Link Token

POST `/plaid/exchange` → Exchange public token for access token

GET `/plaid/accounts` → Get user’s linked bank accounts

GET `/plaid/transactions` → Fetch transactions from a linked bank

🔹 External Integrations:

Auth Service → Ensures only authenticated users can link accounts

Stripe Service → Uses Plaid’s bank verification before enabling payments
