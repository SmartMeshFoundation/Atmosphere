package mpc

import (
	"errors"
	"github.com/BGW-SecureLinearMultiPartyComputation/src/loccs.sjtu.edu.cn/adcrypto/secretshare"
	"math/big"
	"crypto/rand"
)

/**
 * This class implements an BigInt secure multi-party linear function computation.
 *
 * @author 		LoCCS
 * @version		1.0
 */
type LinearMultipartyComputationBigInt struct {
	LinearMultipartyComputation
}

/**
 * Construct linear function MPC scheme with number of participants, threshold and the ID of the participant.
 * <p>
 * The threshold is the max number of semi-honest adversaries, should be less than <i>n</i>/2.
 *
 * @param id ID of this participant.
 * @param participantCount Number of participants.
 * @param threshold Threshold <i>t</i>.
 * @return feedback the constructed  LinearMultipartyComputationBigInt
 * @return error IllegalArgumentException If any of ID, participantCount or threshold is invalid.
 */
func NewLinearMultipartyComputationBigInt(id int, participantCount int, threshold int)(*LinearMultipartyComputationBigInt,error){
	if (participantCount < 3){
		return nil, errors.New("Invalid participant count. Should be larger than 2.")
	}
	if (id < 0 || (id >= participantCount)){
		return nil, errors.New("Invalid id, should be between 0 and participantCount-1")
	}
	if (threshold > (participantCount / 2)){
		return nil, errors.New("Threshold should never greater than 1/2 of the participant count.")
	}
	feedback := new(LinearMultipartyComputationBigInt)
	feedback.id = id
	feedback.participantCount = participantCount
	feedback.threshold = threshold
	feedback.linearMultipartyComputationCalculator = feedback
	feedback.receivedInputs = make([]interface{},participantCount)
	feedback.receivedOutputs = map[int]interface{} {}
	return feedback, nil
}

/**
 * Get a BigInteger Shamir's secret sharing object with the number of participants and the modulus.
 *
 * @param participantCount The number of the participants.
 * @param modulus The modulus <i>p</i>.
 * @return The proper Shamir's secret sharing scheme object.
 * @return error IllegalArgumentException If the number of participants or the modulus is invalid.
 */
func (lmpcb *LinearMultipartyComputationBigInt)getSecretSharing(participantCount int, modulus interface{})(secretshare.
	ShamirSecretSharingInterface,error){
	modulusValue ,ok := modulus.(*big.Int)
	if (!ok) {return nil, errors.New("Invalid type of modulus.")}
	return secretshare.NewShamirSecretSharingBigInt(participantCount,modulusValue)
}

/**
 * Generate a proper BigInteger modulus for Shamir's secret sharing from the coefficients of the linear function and the max value of the secret.
 *
 * @param coefficients The coefficients of the linear function.
 * @param max Max value of a secret.
 * @return The modulus for Shamir's secret sharing scheme.
 * @return error IllegalArgumentException If the coefficients or the max value is invalid.
 */
func (lmpcb *LinearMultipartyComputationBigInt) generateModulus(coefficients []interface{}, max interface{})(interface{},error){
	modulus := big.NewInt(3)
	var err error
	bit := 5      // bit of the prime
	tag := false  // endPrimeGeneratingTag
	pile := big.NewInt(0)  // probably max value of the sum
	for i := 0; i< len(coefficients); i++{
		tmp := big.NewInt(0)
		tmp.Mul(coefficients[i].(*big.Int),max.(*big.Int))
		pile.Add(pile,tmp)
	}
	for (!tag){
		modulus,err = rand.Prime(rand.Reader,bit)
		if (err != nil) {return nil,err}
		tag = modulus.Cmp(pile) > 0
		bit += 5
	}
	return modulus, nil
}

/**
 * Generate output during the output stage.
 *
 * @return The output value.
 */
func (lmpcb *LinearMultipartyComputationBigInt) generateOutputImpl() interface{}{
	modulus := lmpcb.secretSharing.GetModulus()
	pile := big.NewInt(0)
	for i:=0; i < lmpcb.participantCount; i++{
		tmp := big.NewInt(1)
		tmp.Mul(lmpcb.receivedInputs[i].(*big.Int), lmpcb.coefficients[i].(*big.Int))
		tmp.Mod(tmp,modulus.(*big.Int))
		pile.Add(pile,tmp).Mod(pile,modulus.(*big.Int))
	}
	return pile
}
/**
 * Abstract method of checking if the coefficients and the modulus are both valid.
 *
 * @param coefficients Coefficients of the linear function.
 * @param modulus Modulus of the Shamir's scheme.
 * @throws IllegalArgumentException if the coefficients or the modulus is invalid.
 */
func (lmpcb *LinearMultipartyComputationBigInt)checkCoefficientsAndModulus(coefficients []interface{}, modulus interface{}) error{
    if (modulus.(*big.Int).Cmp(big.NewInt(2)) <= 0){
    	return errors.New("Modulus should be greater than 2.")
	}
	if(!modulus.(*big.Int).ProbablyPrime(20)){
    	return errors.New("Modulus is not a Prime.")
	}
	for i := 0; i < len(coefficients); i++{
		if (coefficients[i].(*big.Int).Cmp(big.NewInt(0)) < 0 || coefficients[i].(*big.Int).Cmp(modulus.(*big.Int)) >=0 ){
			return errors.New("One coefficient is too great or too tiny.")
		}
	}
	return nil
}

/**
 * Get BigInteger of value 1.
 *
 * @return BigInteger of value 1.
 */
func (lmpcb *LinearMultipartyComputationBigInt) getElementOne() interface{}{
	return big.NewInt(1)
}
/**
 * Check if the type of input element is valid.
 *
 * @param e Element to be checked.
 * @return True if the type of input element is BigInt, otherwise return false.
 */
func (lmpcb *LinearMultipartyComputationBigInt) checkElement(e interface{}) bool{
	_, ok := e.(*big.Int)
	return ok
}