package poly

/**
 * Abstract class for polynomial over <i>Zp</i> with single variable.
 * <p>
 * The polynomial can be written in the form <i>f</i>(<i>x</i>) = <i>a</i><sub>0</sub> + <i>a</i><sub>1</sub><i>x</i> + ... + <i>a<sub>k</sub></i><i>x<sup>k</sup></i> mod <i>p</i>.
 * While <i>x</i> is the variable, <i>p</i> is the modulus, <i>a</i><sub>0</sub>, <i>a</i><sub>1</sub>, ..., <i>a<sub>k</sub></i> are coefficients in <i>Zp</i>, and <i>k</i> is the degree of the polynomial.
 * <p>
 * The abstract class <code>Polynomial</code> provides default abstract method that calculates result of the polynomial when
 * variable <i>x</i> is given. Subclasses of <code>Polynomial</code> should implement this method.
 *
 * @author 		LoCCS
 * @version		1.0
 */

type Polynomial struct {
	/**
    * degree of the polynomial
    */
	degree int

	/**
	 * coefficients of the polynomial, number of coefficients must be degree + 1.
	 * <p>
	 * coefficients[i] contains <i>a<sub>i</sub></i> (0 ≤ <i>i</i> ≤ <i>k</i>).
	 */
	coefficients []interface{}

	/**
	 * Modulus <i>p</i> of the polynomial.
	 */
	modulus interface{}

    /**
     * Interfaces of Polynomial
    */
	Polynomialcal PolynomialCalculator
}

type PolynomialCalculator interface{
    GetDegree () int
    GetCoefficients() []interface{}
    GetModulus() interface{}
	/**
 	* Abstract method of calculating the results of the polynomial by given value of variable <i>x</i>.
 	*
 	* @param x The value of variable <i>x</i>.
 	* @return The results of the polynomial.
 	* @throws IllegalArgumentException If <i>x</i> is invalid.
    */
	Calculate(x interface{}) (interface{}, error)

}

/**
 * Get the degree of the polynomial.
 *
 * @return degree of the polynomial.
 */
func (poly *Polynomial) GetDegree() int {
	return poly.degree
}

/**
 * Get the coefficients of the polynomial.
 *
 * @return coefficient array.
 */
func (poly *Polynomial) GetCoefficients() []interface{} {
	return poly.coefficients
}

/**
 * Get the modulus of the polynomial.
 *
 * @return modulus of the polynomial.
 */
func (poly *Polynomial) GetModulus() interface{}{
	return poly.modulus
}






