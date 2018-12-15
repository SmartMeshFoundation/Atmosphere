package secretshare

/**
 * The class stores a secret share value in Shamir's secret sharing scheme.
 * <p>
 * A share in Shamir's scheme is a 2-tuple (<i>r</i>, <i>q</i>(<i>r</i>)),
 * where <i>r</i> is the input of the polynomial, <i>q</i>(<i>r</i>) is the result
 * of the polynomial.
 *
 * @author 		LoCCS
 * @version		1.0
 */

type ShamirSecretShareValue struct{
	/**
 	* <i>r</i>, the input of the polynomial.
 	*/
	 r interface{}

	/**
	 * <i>q</i>(<i>r</i>), the result of the polynomial.
	 */
	qr interface{}
}

/**
 * Construct a share value by <i>r</i> and <i>q</i>(<i>r</i>).
 *
 * @param r <i>r</i>, the input of the polynomial.
 * @param qr <i>q</i>(<i>r</i>), the result of the polynomial.
 * @return newShamirSecretShareValue the new constructed ShamirSecretShareValue
 */
func NewShamirSecretShareValue(r interface{}, qr interface{}) (*ShamirSecretShareValue){
	newShamirSecretShareValue := new(ShamirSecretShareValue)
	newShamirSecretShareValue.r = r
	newShamirSecretShareValue.qr = qr
	return newShamirSecretShareValue
}

/**
 * Get <i>r</i>, the input of the polynomial.
 *
 * @return <i>r</i>, the input of the polynomial.
 */
func (sssv *ShamirSecretShareValue) GetR() interface{}{
	return sssv.r
}

/**
 * Get <i>q</i>(<i>r</i>), the result of the polynomial.
 *
 * @return <i>q</i>(<i>r</i>), the result of the polynomial.
 */
func (sssv *ShamirSecretShareValue) GetQr() interface{}{
	return sssv.qr
}