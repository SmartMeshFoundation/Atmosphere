package poly

import (
	"math/big"
	"errors"
	"container/list"
)

/**
 * This class implements an BigInt system of linear equations over <i>Zp</i>.
 *
 * @author 		LoCCS
 * @version		1.0
 */
type LinearEquationSystemBigInt struct {
	LinearEquationSystem
}

/**
 * Construct a system of BigInt linear equation with number of variable count and modulus.
 *
 * @param variableCount Number of variables.
 * @param modulus Modules <i>p</i>
 * @return linearESFeedback The constructed LinearEquationSystemBigInt
 * @return error Whether number of variables or modulus is invalid.
 */
func NewLinearEquationSystemBigInt(variableCount int ,modulus *big.Int)(*LinearEquationSystemBigInt, error){
	linearESFeedback := new(LinearEquationSystemBigInt)
	if (variableCount < 1){
		return nil, errors.New("Number of variables should be greater than 0.");
	}
	if (modulus.Cmp(big.NewInt(2)) <= 0) {
		return nil, errors.New("Modulus should be greater than 2.");
	}
	linearESFeedback.variableCount = variableCount
	linearESFeedback.modulus = modulus
	linearESFeedback.equations = list.New()
	linearESFeedback.LinearEquationSystemITF = linearESFeedback
	return linearESFeedback, nil
}

/**
 * Add a linear equation to the system.
 *
 * @param coefficients Coefficients of the equation.
 * @param constant Constant term of the equation.
 * @return error If coefficients or constant is invalid.
 */
func (les *LinearEquationSystemBigInt) AddEquation(coefficients []interface{}, constant interface{}) error {
	if ((coefficients == nil) || (len(coefficients) != les.variableCount)){
		return errors.New("Number of coefficients should be equals to Number of variables.")
	}
	for i := 0; i < len(coefficients); i++{
		if (!les.checkElement(coefficients[i])) {
			return errors.New("Invalid type of coefficients, should be *big.int.")
		}
	}
	if ((constant == nil) || !les.checkElement(constant)){
		return errors.New("Invalid type of constant, should be *big.int.")
	}

	coefficientsInterface := make([]interface{},les.variableCount)

	for i:=0 ; i < les.variableCount;i++{
		// change every coefficient into zero to modulus-1
		tmp := big.NewInt(0)
		tmp.Mod(coefficients[i].(*big.Int),les.modulus.(*big.Int))
		tmp.Add(tmp,les.modulus.(*big.Int))
		coefficientsInterface[i] = tmp.Mod(tmp,les.modulus.(*big.Int))
	}
	newLinearEquation,err := NewLinearEquation(coefficientsInterface,constant)

	if (err != nil) {
		return err
	} else{
	    les.equations.PushBack(newLinearEquation)
	}
	return nil
}



/**
 * Check if the type of input element is valid.
 *
 * @param e Element to be checked.
 * @return True if the type of input element is big.Int, otherwise return false.
 */
func (les *LinearEquationSystemBigInt) checkElement(e interface{}) bool{
	_,f := e.(*big.Int)
	return f
}

/**
 * Solve system of linear equation.
 *
 * @return The solution if there is single solution exists.
 * And nil if infinitely solutions exist or no solution exists.
 * @return error whether some mistakes happen, such as no solution or infinite solutions
 */
func (les *LinearEquationSystemBigInt) Solve()([]interface{}, error){
	if (les.equations.Len() < les.variableCount) {
		return nil, errors.New("Linear Equations are not enough for solving")
	}

	var i,j,k int
	var validTag bool
	zero := big.NewInt(0) // constant value: zero

	// copy the LinearEquations
	coeffMatrix := make([]interface{},les.variableCount)
	solMatrix := make([]*big.Int,les.variableCount)
	// start from the front of LinearEquations
	countLinearEquations := les.equations.Front()
	for i = 0; i < les.variableCount ; i++{
		if (len(countLinearEquations.Value.(*LinearEquation).coefficients) != les.variableCount){
			return nil, errors.New("Length of Linear Equations doesn't fit the number of Variables")
		}
		tmp := make([]*big.Int, les.variableCount) // one row of coeffMatrix
		for j = 0; j < les.variableCount; j++{
			tmp[j], validTag = countLinearEquations.Value.(*LinearEquation).coefficients[j].(*big.Int)
			if (!validTag) {return nil, errors.New("Invalid type in Linear Equations, should be big.Int")}
		}
		coeffMatrix[i] = tmp
		solMatrix[i] = countLinearEquations.Value.(*LinearEquation).constant.(*big.Int)
		countLinearEquations = countLinearEquations.Next() // get the next LinearEquation
	}

	// Gauss Elimination
	for i = 0;i < les.variableCount; i++{ // i is the row
		// first let mat[i][i] a non-zero number
		if (coeffMatrix[i].([]*big.Int)[i].Cmp(zero) == 0){
			for k= i+1; k < les.variableCount; k++{
				if (coeffMatrix[k].([]*big.Int)[i].Cmp(zero) != 0){
					// change row i and row k
					for j=0;j < les.variableCount;j++{
						tmp := coeffMatrix[k].([]*big.Int)[j]
						coeffMatrix[k].([]*big.Int)[j] = coeffMatrix[i].([]*big.Int)[j]
						coeffMatrix[i].([]*big.Int)[j] = tmp
					}
					// solMatrix should not be forgotten
					tmp := solMatrix[k]
					solMatrix[k] = solMatrix[i]
					solMatrix[i] = tmp
					break
				}
			}
			if k == les.variableCount {
				continue
			} // k == t+1 only if this col has all zero
		}
		inv := big.NewInt(1)
		inv.ModInverse( coeffMatrix[i].([]*big.Int)[i],les.modulus.(*big.Int) ) // inverse of coeff[i][i]
		coeffMatrix[i].([]*big.Int)[i].Mul(coeffMatrix[i].([]*big.Int)[i],inv).
			Mod(coeffMatrix[i].([]*big.Int)[i],les.modulus.(*big.Int)) // make coeff[i][i] = 1
		// for row i , multiply every number with inv
		for j = i+1; j < les.variableCount; j++{
			coeffMatrix[i].([]*big.Int)[j].Mul(coeffMatrix[i].([]*big.Int)[j],inv).
				Mod(coeffMatrix[i].([]*big.Int)[j],les.modulus.(*big.Int))
		}
		// solMatrix of row i should not be forgotten
		solMatrix[i].Mul(solMatrix[i],inv)
		solMatrix[i].Mod(solMatrix[i],les.modulus.(*big.Int))
		// for row below i, substract to make coe[k][i] = 0
		for k= i+1; k < les.variableCount; k++ {
			multiple := big.NewInt(1) // the multiple of row k substract
			multiple.Set(coeffMatrix[k].([]*big.Int)[i])
			for j = i; j < les.variableCount; j++{
				tmp := big.NewInt(1)
				tmp.Mul(multiple,coeffMatrix[i].([]*big.Int)[j]).Mod(tmp,les.modulus.(*big.Int))
				coeffMatrix[k].([]*big.Int)[j].Sub(coeffMatrix[k].([]*big.Int)[j],tmp).
					Mod(coeffMatrix[k].([]*big.Int)[j],les.modulus.(*big.Int))
			}
			// solMatrix should also be refreshed
			tmp := big.NewInt(1)
			tmp = tmp.Mul(multiple,solMatrix[i]).Mod(tmp,les.modulus.(*big.Int))
			solMatrix[k] = solMatrix[k].Sub(solMatrix[k],tmp).Mod(solMatrix[k],les.modulus.(*big.Int))
		}
	}

	// judge whether the equation system has zero/infinite solutions
	for i =0 ; i < les.variableCount; i++{
		// if not all numbers in the last row is azero
		if (coeffMatrix[les.variableCount-1].([]*big.Int)[i].Cmp(zero) != 0) {break;}
		// all-zero processing
		if (i== les.variableCount-1 && solMatrix[les.variableCount-1].Cmp(zero) == 0){
			return nil, errors.New("Infinite solutions for this LinearEquationSystem.")
		}
		if (i== les.variableCount-1 && solMatrix[les.variableCount-1].Cmp(zero) != 0){
			return nil, errors.New("No solutions for this LinearEquationSystem.")
		}
	}

	//fmt.Println(solMatrix)
	feedback := make([]interface{},les.variableCount) // the solution vector
	feedback[les.variableCount - 1] = solMatrix[les.variableCount - 1]
	// back calculation from the t+1 row
	for i = les.variableCount - 2; i >= 0; i-- {
		tmp := big.NewInt(0)
		for j = i + 1; j < les.variableCount; j++{
			tmp2 := big.NewInt(1)
			tmp2 = tmp2.Mul(feedback[j].(*big.Int),coeffMatrix[i].([]*big.Int)[j])
			tmp = tmp.Add(tmp,tmp2)
		}
		tmp3 := big.NewInt(0)
		tmp3 = tmp3.Sub(solMatrix[i],tmp).Mod(tmp3,les.modulus.(*big.Int))
		inv2 := big.NewInt(1)
		inv2 = inv2.ModInverse(coeffMatrix[i].([]*big.Int)[i],les.modulus.(*big.Int))
		feedback[i] = tmp3.Mul(tmp3,inv2)
		feedback[i] = feedback[i].(*big.Int).Mod(feedback[i].(*big.Int),les.modulus.(*big.Int))
	}
	return feedback, nil
}
