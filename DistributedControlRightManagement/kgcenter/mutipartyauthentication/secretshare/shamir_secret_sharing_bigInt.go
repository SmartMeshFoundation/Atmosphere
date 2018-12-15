package secretshare

import (
	"math/big"
	"errors"
	"crypto/rand"
	"github.com/BGW-SecureLinearMultiPartyComputation/src/loccs.sjtu.edu.cn/adcrypto/poly"
)

/**
 * The class implements Shamir's secret sharing scheme on BigInt field.
 *
 * @author 		LoCCS
 * @version		1.0
 */

type ShamirSecretSharingBigInt struct {
	ShamirSecretSharing
}

/**
 * Construct secret sharing scheme with the number of participants and BigInt modulus.
 *
 * @param participantCount The number of participants that share the secret.
 * @param modulus The order the finite field used by the polynomial.
 * @return feedback the newly constructed ShamirSecretSharingBigInt
 * @return error If the number of participants or modulus is invalid.
 */
func NewShamirSecretSharingBigInt(participantCount int,modulus *big.Int) (*ShamirSecretSharingBigInt, error){
	if (participantCount < 2){
		return nil, errors.New("Invalid participant count. Should be larger than 1.")
	}
	if (modulus.Cmp(big.NewInt(2)) <= 0){
		return nil, errors.New("Invalid modulus. Should be larger than 2.")
	} else if( !modulus.ProbablyPrime(20)){
		return nil, errors.New("Invalid modulus. Should be prime")
	}
	feedback := new(ShamirSecretSharingBigInt)
	feedback.participantCount = participantCount
	feedback.modolus = modulus
	feedback.ShamirSecretSharingITF = feedback
	feedback.SecretSharingSchemeITF = &feedback.ShamirSecretSharing
	return feedback, nil
}

/**
 * Generate<i>n</i> BigInteger random auxiliary data from each participant.
 *
 * @return Random auxiliary
 */
func (sssb *ShamirSecretSharingBigInt) GenerateRandomAuxiliary() []interface{}{
     feedback := make([]interface{}, sssb.participantCount)
	 for i := 0 ; i < sssb.participantCount; i++{
	 	 tag := false
	 	 for (!tag){
	 	 	 tmp, _ := rand.Int(rand.Reader,sssb.modolus.(*big.Int))
	 	 	 tag = (tmp.Cmp(big.NewInt(0)) > 0)
			 feedback[i] = tmp
		 }
	 }
	 return feedback
}

/**
 * Get a random <i>k</i>-1 degree polynomial over <i>Zp</i>.
 * <p>
 * <i>a</i><sub>0</sub> is the BigInt secret specified by input parameter,
 * and <i>a</i><sub>1</sub>, <i>a</i><sub>2</sub>, ..., <i>a</i><sub><i>k</i>-1</sub>
 * are chosen randomly in <i>Zp</i>.
 *
 * @param a0 Constant term of the polynomial.
 * @return The polynomial object.
 */
func (sssb *ShamirSecretSharingBigInt) GetRandomPolynomial(a0 interface{}) poly.PolynomialCalculator{
	degree := sssb.access.(*ThresholdAccessStructure).GetThreshold() - 1
	coeff := make([]*big.Int,degree + 1)
    coeff[0] = a0.(*big.Int)
	for i := 1 ; i < degree+1; i++{
		tag := false
		for (!tag){
			tmp, _ := rand.Int(rand.Reader,sssb.modolus.(*big.Int))
			tag = (tmp.Cmp(big.NewInt(0)) > 0)
			coeff[i] = tmp
		}
	}
	feedback, _ := poly.NewPolynomialBigInt(degree,coeff,sssb.modolus.(*big.Int))
	return feedback
}

/**
 * Create default auxiliary data (BigInt array) from generate shares.
 * <p>
 * i.e. 1, 2, ..., <i>n</i>
 *
 * @return Default auxiliary data
 */
func (sssb *ShamirSecretSharingBigInt) CreateDefaultAuxiliary() []interface{}{
	feedback := make([]interface{}, sssb.participantCount)
	for i := 0 ; i < sssb.participantCount; i++{
		feedback[i] = big.NewInt(int64(i+1))
	}
	return feedback
}

/**
 * Get a system linear equation with <i>k</i> variables over <i>Zp</i> for calculating secret.
 *
 * @return Linear equation system object(LinearEquationSystemBigInt).
 */
func (sssb *ShamirSecretSharingBigInt) GetEquationSystem() poly.LinearEquationSystemCalculator{
	variableCount := sssb.access.(*ThresholdAccessStructure).GetThreshold()
	feedback, _ := poly.NewLinearEquationSystemBigInt(variableCount,sssb.modolus.(*big.Int))
	return feedback
}

/**
 * Restore the coefficients of the equation by the first element of the share.
 * <p>
 * i.e. 1, <i>x</i>, <i>x</i><sup>2</sup>, ... , <i>x</i><sup><i>k</i>-1</sup> mod <i>p</i>
 *
 * @param x The first element of the share.
 * @return The coefficient array(BigInt array).
 */
func (sssb *ShamirSecretSharingBigInt) GetEquationCoefficients(x interface{}) []interface{}{
	threshold := sssb.access.(*ThresholdAccessStructure).GetThreshold()
	feedback := make([]interface{},threshold)
	pile := big.NewInt(1)
	for i := 0; i < threshold; i++{
		tmp := big.NewInt(1)
		tmp.Set(pile)
		feedback[i] = tmp
		pile.Mul(x.(*big.Int), pile).Mod(pile ,sssb.modolus.(*big.Int))
	}
	return feedback
}

/**
 * Check if the type of input element is valid.
 *
 * @param e Element to be checked.
 * @return True if the type of input element is bigInt, otherwise return false.
 */
func (sssb *ShamirSecretSharingBigInt) checkElement(e interface{}) bool{
	_, ok := e.(*big.Int)
	return ok
}



