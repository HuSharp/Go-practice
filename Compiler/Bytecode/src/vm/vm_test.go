package vm

import (
	"fmt"
	"src/ast"
	"src/compiler"
	"src/lexer"
	"src/object"
	"src/parser"
	"testing"
)

type vmTestCase struct {
	input    string
	expected interface{}
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"1 * 2", 2},
		{"4 / 2", 2},
		{"50 / 2 * 2 + 10 - 5", 55},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"5 * (2 + 10)", 60},
		{"-5", -5},
		{"-10", -10},
		{"-50 + 100 + -50", 0},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}
	runVmTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`"monkey"`, "monkey"},
		{`"mon" + "key"`, "monkey"},
		{`"mon" + "key" + "banana"`, "monkeybanana"},
	}
	runVmTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"[]", []int{}},
		{"[1, 2, 3]", []int{1, 2, 3}},
		{"[1 + 2, 3 * 4, 5 + 6]", []int{3, 12, 11}},
	}
	runVmTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []vmTestCase{
		{
			"{}", map[object.HashKey]int64{},
		},
		{
			"{1: 2, 2: 3}",
			map[object.HashKey]int64{
				(&object.Integer{Value: 1}).HashKey(): 2,
				(&object.Integer{Value: 2}).HashKey(): 3,
			},
		},
		{
			"{1 + 1: 2 * 2, 3 + 3: 4 * 4}",
			map[object.HashKey]int64{
				(&object.Integer{Value: 2}).HashKey(): 4,
				(&object.Integer{Value: 6}).HashKey(): 16,
			},
		},
	}
	runVmTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"!(if (false) { 5; })", true},
	}
	runVmTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][0 + 2]", 3},
		{"[[1, 1, 1]][0][0]", 1},
		{"[][0]", Null},
		{"[1, 2, 3][99]", Null},
		{"[1][-1]", Null},
		{"{1: 1, 2: 2}[1]", 1},
		{"{1: 1, 2: 2}[2]", 2},
		{"{1: 1}[0]", Null},
		{"{}[0]", Null},
		//{"{"one": 1}"},
	}
	runVmTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{"if (true) { 10 }", 10},
		{"if (true) { 10 } else { 20 }", 10},
		{"if (false) { 10 } else { 20 } ", 20},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 > 2) { 10 }", Null},
		{"if (false) { 10 }", Null},
		{"if ((if (false) { 10 })) { 10 } else { 20 }", 20},
	}
	runVmTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []vmTestCase{
		{"let one = 1; one", 1},
		{"let one = 1; let two = 2; one + two", 3},
		{"let one = 1; let two = one + one; one + two", 3},
	}
	runVmTests(t, tests)
}

func TestCallingFunctionsWithoutArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
		let fivePlusTen = fn() { 5 + 10; };
		fivePlusTen();
		`,
			expected: 15,
		},
		{
			input: `
		let one = fn() { 1; };
		let two = fn() { 2; };
		one() + two()
		`,
			expected: 3,
		},
		{
			input: `
		let a = fn() { 1 };
		let b = fn() { a() + 1 };
		let c = fn() { b() + 1 };
		c();
		`,
			expected: 3,
		},
	}
	runVmTests(t, tests)
}

func TestFunctionsWithReturnStatement(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
		let earlyExit = fn() { return 99; 100; };
		earlyExit();
		`,
			expected: 99,
		},
		{
			input: `
		let earlyExit = fn() { return 99; return 100; };
		earlyExit();
		`,
			expected: 99,
		},
	}
	runVmTests(t, tests)
}

func TestFunctionsWithoutReturnValue(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
		let noReturn = fn() { };
		noReturn();
		`,
			expected: Null,
		},
		{
			input: `
		let noReturn = fn() { };
		let noReturnTwo = fn() { noReturn(); };
		noReturn();
		noReturnTwo();
		`,
			expected: Null,
		},
	}
	runVmTests(t, tests)
}

func TestFirstClassFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
		let returnsOne = fn() { 1; };
		let returnsOneReturner = fn() { returnsOne; };
		returnsOneReturner()();
		`,
			expected: 1,
		},
		{
			input: `
		let returnsOneReturner = fn() {
			let returnsOne = fn() { 1; };
			returnsOne;
		};
		returnsOneReturner()();
		`,
			expected: 1,
		},
	}
	runVmTests(t, tests)
}

func TestCallingFunctionsWithBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
		let one = fn() { let one = 1; one };
		one();
		`,
			expected: 1,
		},
		{
			input: `
		let oneAndTwo = fn() { let one = 1; let two = 2; one + two; };
		oneAndTwo();
		`,
			expected: 3,
		},
		{
			input: `
		let oneAndTwo = fn() { let one = 1; let two = 2; one + two; };
		let threeAndFour = fn() { let three = 3; let four = 4; three + four; };
		oneAndTwo() + threeAndFour();
		`,
			expected: 10,
		},
		{
			input: `
		let firstFoobar = fn() { let foobar = 50; foobar; };
		let secondFoobar = fn() { let foobar = 100; foobar; };
		firstFoobar() + secondFoobar();
		`,
			expected: 150,
		},
		{
			input: `
		let globalSeed = 50;
		let minusOne = fn() {
			let num = 1;
			globalSeed - num;
		}
		let minusTwo = fn() {
			let num = 2;
			globalSeed - num;
		}
		minusOne() + minusTwo();
		`,
			expected: 97,
		},
	}
	runVmTests(t, tests)
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()
	for _, tt := range tests {
		program := parse(tt.input)

		com := compiler.New()
		if err := com.Compile(program); err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(com.Bytecode())
		if err := vm.Run(); err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.LastPoppedStackElem()
		testExpectedObject(t, tt.expected, stackElem)
	}
}

func testExpectedObject(t *testing.T,
	expected interface{},
	actual object.Object) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		if err := testIntegerObject(int64(expected), actual); err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case string:
		if err := testStringObject(expected, actual); err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case bool:
		if err := testBooleanObject(expected, actual); err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case []int:
		if err := testArrayObject(expected, actual); err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case map[object.HashKey]int64:
		if err := testHashObject(expected, actual); err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case *object.NULL:
		if actual != Null {
			t.Errorf("object is not Null: %T (%+v)", actual, actual)
		}
	}
}

func testIntegerObject(expected int64, obj object.Object) error {
	result, ok := obj.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)",
			obj, obj)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%d, want=%d",
			result.Value, expected)
	}
	return nil
}

func testStringObject(expected string, obj object.Object) error {
	result, ok := obj.(*object.String)
	if !ok {
		return fmt.Errorf("object is not String. got=%T (%+v)",
			obj, obj)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%q, want=%q",
			result.Value, expected)
	}
	return nil
}

func testBooleanObject(expected bool, obj object.Object) error {
	result, ok := obj.(*object.Boolean)
	if !ok {
		return fmt.Errorf("object is not Boolean. got=%T (%+v)",
			obj, obj)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%t, want=%t",
			result.Value, expected)
	}
	return nil
}

func testArrayObject(expected []int, obj object.Object) error {
	array, ok := obj.(*object.Array)
	if !ok {
		return fmt.Errorf("object is not Array. got=%T (%+v)",
			obj, obj)
	}

	if len(array.Elements) != len(expected) {
		return fmt.Errorf("wrong num of elements. got=%d, want=%d",
			len(array.Elements), len(expected))
	}

	for i, expectedElem := range expected {
		if err := testIntegerObject(int64(expectedElem), array.Elements[i]); err != nil {
			return err
		}
	}
	return nil
}

func testHashObject(expected map[object.HashKey]int64, obj object.Object) error {
	hash, ok := obj.(*object.Hash)
	if !ok {
		return fmt.Errorf("object is not Hash. got=%T (%+v)",
			obj, obj)
	}

	if len(hash.Pairs) != len(expected) {
		return fmt.Errorf("wrong num of elements. got=%d, want=%d",
			len(hash.Pairs), len(expected))
	}

	for expectedKey, expectedValue := range expected {
		pair, exist := hash.Pairs[expectedKey]
		if !exist {
			return fmt.Errorf("no pair for given key in Pairs. got key=%+v", expectedKey)
		}
		if err := testIntegerObject(expectedValue, pair.Value); err != nil {
			return err
		}
	}
	return nil
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}
