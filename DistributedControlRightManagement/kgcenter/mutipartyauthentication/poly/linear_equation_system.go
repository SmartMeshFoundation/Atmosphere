package poly

import (
	"container/list"
)

/**
 * Abstract class for a system of linear equations over <i>Zp</i>.
 * <p>
 * A general system of <i>m</i> linear equations with <i>n</i> variables can be written as:
 * <p>
 * &nbsp;&nbsp;&nbsp;&nbsp;<i>a</i><sub>11</sub><i>x</i><sub>1</sub> + <i>a</i><sub>12</sub><i>x</i><sub>2</sub> + ... + <i>a</i><sub>1<i>n</i></sub><i>x<sub>n</sub></i> = <i>b</i><sub>1</sub>
 * <p>
 * &nbsp;&nbsp;&nbsp;&nbsp;<i>a</i><sub>21</sub><i>x</i><sub>1</sub> + <i>a</i><sub>22</sub><i>x</i><sub>2</sub> + ... + <i>a</i><sub>2<i>n</i></sub><i>x<sub>n</sub></i> = <i>b</i><sub>2</sub>
 * <p>
 * &nbsp;&nbsp;&nbsp;&nbsp;......
 * <p>
 * &nbsp;&nbsp;&nbsp;&nbsp;<i>a</i><sub><i>m</i>1</sub><i>x</i><sub>1</sub> + <i>a</i><sub><i>m</i>2</sub><i>x</i><sub>2</sub> + ... + <i>a<sub>mn</sub></i><i>x<sub>n</sub></i> = <i>b<sub>m</sub></i>
 * <p>
 * where <i>x</i><sub>1</sub>, <i>x</i><sub>2</sub>, ..., <i>x<sub>n</sub></i> are the variables, <i>a</i><sub>11</sub>, <i>a</i><sub>12</sub>, <i>a<sub>mn</sub></i> are the coefficients of the system,
 * and <i>b</i><sub>1</sub>, <i>b</i><sub>2</sub>, ..., <i>b<sub>m</sub></i> are the constant terms.
 * <p>
 * A solution of a linear system is an assignment of values to the variables <i>x</i><sub>1</sub>, <i>x</i><sub>2</sub>, ..., <i>x<sub>n</sub></i>
 * such that each of the equations is satisfied. A linear system may behave in any one of three possible ways:
 * <ol>
 * <li>The system has infinitely many solutions.
 * <li>The system has a single unique solution.
 * <li>The system has no solution.
 * </ol>
 * The abstract class <code>LinearEquationSystem</code> provides default abstract method that
 * calculate the solution of the linear equation set when there is single solution exists.
 * Subclasses of <code>LinearEquationSystem</code> should implement this method.
 *
 * @author 		LoCCS
 * @version		1.0
 */
type LinearEquationSystem struct {

	/**
    * Number of variables.
    */
	variableCount int

	/**
	 * Modulus <i>p</i>.
	 */
	modulus interface{}

	/**
	 * List that store all equations in the system.
	 */
	equations *list.List

	/**
    * Abstract Interfaces of LinearEquationSystem
	*/
	LinearEquationSystemITF LinearEquationSystemCalculator
}

type LinearEquationSystemCalculator interface {
	/**
	* Clear all equations in the system.
    */
	Clear()

	/**
    * Abstract method of solving system of linear equation.
    *
    * @return The solution if there is single solution exists.
    * And null if infinitely solutions exist or no solution exists.
    */
    Solve() ([]interface{},error)

	/**
    * Abstract method of checking if the type of input element is valid.
    *
    * @param e Element to be checked.
    * @return True if the type of input element is valid, otherwise return false.
    */
    checkElement(e interface{}) bool

	/**
	 * Add a linear equation to the system.
	 *
	 * @param coefficients Coefficients of the equation.
	 * @param constant Constant term of the equation.
	 * @throws IllegalArgumentException If coefficients or constant is invalid.
	 */
    AddEquation (coefficients []interface{}, constant interface{}) error
}

func (lqs *LinearEquationSystem) Clear(){
    lqs.equations.Init()
}