# bookmark-management
Learning golang backend engineer by doing

- Khởi tạo project: go mod init github.com/jaimesHub/bookmark-management -> `go.mod` -> add git
- go mod init
- project folder structure: go clean architecture 
  - cmd : contains golang command lines
  - internal : things relate to application
    - api -> focus
    - handler -> focus
    - model
    - service
    - api
    - repository -> focus 
  - practice approach from bottom to up