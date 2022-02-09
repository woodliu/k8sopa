package validate

type Validation interface {
    Validate([]byte)string
    Dump() //TODO
}