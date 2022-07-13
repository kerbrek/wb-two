Что выведет программа? Объяснить вывод программы.

```go
package main

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

func test() *customError {
	{
		// do something
	}
	return nil
}

func main() {
	var err error
	err = test()
	if err != nil {
		println("error")
		return
	}
	println("ok")
}
```

Ответ:
```
Выведет:
  error
Созданной интерфейсной переменной err мы назначаем динамический тип *customError
и присваиваем значение типа равное nil. Чтобы проходить проверку на равенство с
nil динамический тип интерфейсного значения тоже должен быть nil.
```
