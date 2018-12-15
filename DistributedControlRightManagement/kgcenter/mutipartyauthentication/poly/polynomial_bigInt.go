package poly

import (
	"math/big"
	"errors"
)

/**
 * This class implements an math/big polynomial over <i>Zp</i> with single variable.
 *
 * @author 		LoCCS
 * @version		1.0
 */
type PolynomialBigInt struct{
 	Polynomial
}

/**
 * Construct polynomialBigInt with degree, coefficients and modulus.
 * <p>
 * Number of coefficients must be degree + 1, coefficients[i] contains <i>a<sub>i</sub></i> (0 ≤ <i>i</i> ≤ <i>k</i>).
 *
 * @param degree Degree of the polynomial.
 * @param coefficients	Coefficients of the polynomial.
 * @param modulus Modulus p of the polynomial.
 * @return polyFeedback The constructed polynomialBigInt
 * @return error The error tag of whether an error happens
 */
func NewPolynomialBigInt(degree int, coefficients []*big.Int, modulus *big.Int)(*PolynomialBigInt,error){
	polyFeedback := new(PolynomialBigInt)
	if (degree < 0) {
		return nil , errors.New("Invalid polynomial degree, should not be less than 0.")
	}
	if ((coefficients == nil) || len(coefficients) != (degree + 1)){
		return nil , errors.New("Number of polynomial coefficients should be degree + 1.")
	}
	if (modulus == nil){
		return nil , errors.New("Polynomial(of crypto) should be over Zp.")
	} else if (modulus.Cmp(big.NewInt(2)) < 0){
		return nil , errors.New("Invalid polynomial modulus, should be greater than 1.")
	}
	polyFeedback.degree = degree
	polyFeedback.modulus = modulus
    polyFeedback.coefficients = make([]interface{},degree+1)

	for i:=0; i < degree+1; i++{
		if (coefficients[i].Cmp(big.NewInt(0)) < 0 || coefficients[i].Cmp(modulus) >= 0 ){
			tmp := big.NewInt(0)
			polyFeedback.coefficients[i] = tmp.Mod(coefficients[i], modulus)
		} else{
			polyFeedback.coefficients[i] = coefficients[i]
		}
	}
	polyFeedback.Polynomialcal = polyFeedback
	return polyFeedback , nil
}

/**
 * Calculate the results of the polynomial by given value of variable <i>x</i>.
 *
 * @param x The value of variable <i>x</i>.
 * @return The results of the polynomial.
 * @throws IllegalArgumentException If <i>x</i> is invalid.
 */
func (poly *PolynomialBigInt) Calculate(x interface{}) (interface{}, error){
	xValue, ok := x.(*big.Int)
	if (!ok){
		return nil , errors.New("Invalid type of input, should be big.Int.")
	}
	feedback := big.NewInt(0)
	pile := big.NewInt(1)
	tmp := big.NewInt(1)
	for i := 0; i < poly.degree + 1; i++{
		tmp.Mul(poly.coefficients[i].(*big.Int), pile).Mod(tmp,poly.modulus.(*big.Int))
		feedback.Add(feedback,tmp).Mod(feedback,poly.modulus.(*big.Int))
		pile.Mul(pile,xValue).Mod(pile,poly.modulus.(*big.Int))
	}
	return feedback, nil
}



