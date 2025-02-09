# Boonkosang Construction Management Backend

### Developed by Beer Do San

---

## Overview

The Boonkosang Construction Management Backend is designed to handle essential documents and processes for construction projects, covering **BOQ (Bill of Quantities)**, **Quotations**, **Invoices**, and **Contracts**. This system streamlines workflows by enabling the creation and management of these documents, supporting efficiency and accuracy throughout project operations.

## Features

- **BOQ Management:** Manage BOQs to specify tasks and required materials per project, including detailed cost estimation.
- **Quotation Management:** Generate and approve quotations based on pre-approved BOQs, complete with tax calculations and net amounts.
- **Invoice Generation:** Generate invoices for initiated projects and track customer payment status.
- **Contract Management:** Create and store contracts to ensure compliance with construction project agreements.
- **Export & Reporting:** Supports data export in JSON format for easy document storage and readability.

## Tech Stack

- **Language:** Go
- **Database:** PostgreSQL

## Getting Started

1. **Clone the Repository:**

   ```sh
   git clone https://github.com/beerth21624/boonkosang-construction-be.git
   ```

2. **Install Dependencies:**
   Run the following to install required packages:

   ```sh
   go mod tidy
   ```

3. **Setup Database:**
   Create and configure the SQL database, then add connection settings to your environment file.

4. **Run the Application:**
   ```sh
   cd cmd/api
   go run main.go
   ```

## License

MIT License.
