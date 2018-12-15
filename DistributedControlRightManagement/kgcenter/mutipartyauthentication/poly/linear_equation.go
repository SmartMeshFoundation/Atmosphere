package poly
/**
 *
 * This class implements a single equation of an Linear Equation System
 * @author 		LoCCS
 * @version		1.0
 */
type LinearEquation struct {
	/**
    * Coefficients of the single equation.
    */
	coefficients []interface{}

	/**
	 * Constant term of the single equation.
	 */
	constant interface{}
}

/**
 * Construct linear equation with coefficients and constant term.
 *
 * @param coefficients Coefficients of the equation.
 * @param constant Constant term of the equation.
 * @return lq the constrcted LinearEquation
 * @return error whether an error happens when constructing LinearEquation
 */
func NewLinearEquation(coefficient []interface{}, constant interface{})(*LinearEquation,error){
	lq := new(LinearEquation)
	lq.constant = constant
	lq.coefficients = make([]interface{}, len(coefficient))
	for i:=0; i < len(coefficient); i++{
		lq.coefficients[i] = coefficient[i]
	}
	return lq,nil
}


/**
* Get the coefficients.
*
* @return Coefficient array.
*/
func (lq *LinearEquation) GetCoefficients() []interface{} {
    return lq.coefficients;
}

/**
 * Get the constant term.
 *
 * @return Constant term.
 */
func (lq *LinearEquation) GetConstant() interface{} {
	return lq.constant;
}