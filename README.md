# Book Review System

Welcome to the Book Review System project! This system allows users to log in, review books, and perform other related actions.

## Features
- User authentication: Users can register, log in, and log out securely.
- Book review: Users can read and write reviews for books.
- Book rating: Users can rate books and see the average rating.
- Book search: Users can search for books by title, author, or genre.
- User profile: Users can view and update their few profile information.

## Technologies Used
- GoLang: Backend development
- HTML/CSS/JavaScript: Frontend development
- MySQL: Database management
<!-- - JWT (JSON Web Tokens): Authentication mechanism -->

## Getting Started

### Prerequisites
- GoLang installed on your system. You can download it from [here](https://golang.org/dl/).
- PostgreSQL installed on your system. You can download it from [here](https://www.postgresql.org/download/).
- Node.js and npm installed on your system for frontend development.

### Installation
1. Clone the repository:
    ```bash
    git clone github.com/iamgak/go-bookreview
    ```

2. Navigate to the project directory:
    ```bash
    cd go-bookreview
    ```

3. Set up the database:
    - Create a MySQL database named `go-bookreview`.
    - Import the database schema from `database/schema.sql`.


4. Install dependencies:
    ```bash
    go mod tidy
    ```

5. Run the server:
    ```bash
    go run cmd/cli
    ```
    optional if you want to change port number and root info
    ```bash

    go run cmd/cli -addr=":8000" -dsn=""root:@/bookstore?parseTime=true"
    ```

5. Access the application:
    Open your web browser and navigate to `http://localhost:8000`.

### Usage
- Register a new user account By PostMethod `http://localhost:8000/user/register`.
- Login in with existing credentials By PostMethod `http://localhost:8000/user/login`.
- Browse all reviews By GetMethod `http://localhost:8000/book/listing`.
- Delete reviews By GetMethod `http://localhost:8000/book/delete/id`.
- Browse reviews By ISBN or Author or Book GetMethod `http://localhost:8000/book/isbn/934-3434` or `http://localhost:8000/book/author/murakami` .
- Write reviews for books you've read By PostMethod After Login `http://localhost:8000/book/create` .
- Browse reviews By GetMethod
- Request Forget Password By PostMethod  `http://localhost:8000/user/forget_password/`token send on given email if registered but in this case you will copy it from table forget_passw during new password .
- Change Forget Password By PostMethod  `http://localhost:8000/user/new_password/token` token from forget_password .

## Contributing
Contributions are welcome! If you'd like to contribute to this project, please fork the repository and submit a pull request with your changes.

