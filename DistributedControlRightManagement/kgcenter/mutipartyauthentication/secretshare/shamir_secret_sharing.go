package secretshare

import (
	"github.com/BGW-SecureLinearMultiPartyComputation/src/loccs.sjtu.edu.cn/adcrypto/poly"
	"errors"
	"fmt"
)

/**
 * Abstract class for shamir's secret sharing scheme.
 * <p>
 * The abstract class <code>ShamirSecretSharing</code> generalize Shamir's secret sharing scheme in
 * "Shamir A. How to share a secret. Communications of the ACM. 1979 Nov 1;22(11):612-3."
 * <p>
 * Shamir's secret sharing scheme is a (<i>k</i>, <i>n</i>) threshold scheme that:
 * <ol>
 * 		<li> <i>n</i> participants;
 * 		<li> Knowledge of any <i>k</i> or more shares make secret easily computable;
 * 		<li> Knowledge of any <i>k</i>-1 or fewer shares leaves secret completely undetermined.
 * </ol>
 * <code>ShamirSecretSharing</code> use <code>ThresholdAccessStructure</code> as its internal access structure.
 * <p>
 * The abstract class <code>ShamirSecretSharing</code> uses a <i>k</i>-1 degree polynomial over <i>Zp</i>, and provides
 * default methods to generate shares and calculate secret which use some auxiliary abstract methods.
 * Subclasses of <code>ShamirSecretSharing</code> should implement these auxiliary methods.
 *
 * @author 		LoCCS
 * @version		1.0
 */

type ShamirSecretSharing struct {
	// father strcut
	SecretSharingScheme

	/**
    * Modulus <i>p</i>, must be prime.
    */
	modolus interface{}

	/**
   * Abstract Interfaces of ShamirSecretSharing
   */
	ShamirSecretSharingITF ShamirSecretSharingInterface
}

type ShamirSecretSharingInterface interface {

	GetModulus() interface{}

	IsInitialized() bool

	generateSharesImpl(secret interface{}, auxiliary []interface{}) ([]*SecretShare, error)

	calculateSecretImpl(shares []*SecretShare) (interface{},error)

	SetAccessStructure(access AccessStructureInterface) error

	GetParticipantCount() int

	GetAccessStructure() AccessStructureInterface

	GenerateShares(secret interface{}, auxiliary []interface{}) ([]*SecretShare, error)

	CalculateSecret(shares []*SecretShare) (interface{}, error)

	/**
    * Abstract method of generating <i>n</i> random auxiliary data from each participant.
    *
    * @return Random auxiliary
    */
	GenerateRandomAuxiliary() []interface{}

	/**
 	* Abstract method of getting a random <i>k</i>-1 degree polynomial over <i>Zp</i>.
 	* <p>
 	* <i>a</i><sub>0</sub> is the secret specified by input parameter,
 	* and <i>a</i><sub>1</sub>, <i>a</i><sub>2</sub>, ..., <i>a</i><sub><i>k</i>-1</sub>
 	* are chosen randomly in <i>Zp</i>.
	 *
 	* @param a0 Constant term of the polynomial.
 	* @return The polynomial object.
 	*/
	GetRandomPolynomial(a0 interface{}) poly.PolynomialCalculator

	/**
 	* Abstract method of creating default auxiliary data from generate shares.
 	* <p>
 	* i.e. 1, 2, ..., <i>n</i>
 	*
	* @return Default auxiliary data
	*/
    CreateDefaultAuxiliary() []interface{}

	/**
 	* Abstract method of getting a system linear equation with <i>k</i> variables over <i>Zp</i> for calculating secret.
 	*
	 * @return Linear equation system object.
 	*/
 	GetEquationSystem() poly.LinearEquationSystemCalculator

	/**
 	* Abstract method of restoring the coefficients of the equation by the first element of the share.
 	* <p>
 	* i.e. 1, <i>x</i>, <i>x</i><sup>2</sup>, ... , <i>x</i><sup><i>k</i>-1</sup> mod <i>p</i>
 	*
 	* @param x The first element of the share.
 	* @return The coefficient array.
 	*/
	GetEquationCoefficients(x interface{}) []interface{}

   /**
   * Abstract method of checking if the type of input element is valid.
   *
   * @param e Element to be checked.
   * @return True if the type of input element is valid, otherwise return false.
   */
    checkElement(e interface{}) bool
}

/**
 * Get the modulus.
 *
 * @return Modulus <i>p</i>.
 */
func (sss *ShamirSecretSharing) GetModulus() interface{}{
	return sss.modolus
}

/**
 * Determine if the scheme object is initialized properly for generating shares and calculating secret.
 * <p>
 * Scheme object is ready when following elements are set: 1) access structure, 2) modulus.
 *
 * @return True if the scheme object is initialized properly, otherwise return false.
 */
func (sss *ShamirSecretSharing) IsInitialized() bool{
	if (sss.access == nil || sss.modolus == nil){
		fmt.Println(sss.access)
		fmt.Println(sss.modolus)
		return false
	} else{ return true }
}

/**
 * Generate shares from input secret under Shamir's secret sharing scheme.
 * <p>
 * The auxiliary data are <i>n</i> different numbers from <i>Zp</i><sup>*</sup>, each
 * for one participant to calculate its share. Note that 0 cannot be used since it will
 * disclose the secret immediately.
 * <p>
 * If parameter auxiliary is null, then number 1, 2, ..., <i>n</i> are used
 * according to the original paper.
 *
 * @param secret The secret from which shares are generated.
 * @param auxiliary Auxiliary data for generating shares. Can be nil(use default auxiliary).
 * @return N shares that generated from the input secret, one for each participant.
 * @return error If the input secret is invalid.
 */
func (sss *ShamirSecretSharing) generateSharesImpl(secret interface{}, auxiliary []interface{}) ([]*SecretShare, error){
    if (!sss.ShamirSecretSharingITF.checkElement(secret)){
    	return nil, errors.New("Invalid type of secret.")
	}

	// use default auxiliary if nil
	if (auxiliary == nil){
		auxiliary = sss.ShamirSecretSharingITF.CreateDefaultAuxiliary()
	} else { // check whether auxiliary is valid
		if (len(auxiliary) != sss.participantCount){
			return nil, errors.New("Invalid number of auxiliary data, should be equal to number of participants.")
		}
		for i:=0; i<len(auxiliary); i++{
			if (!sss.ShamirSecretSharingITF.checkElement(auxiliary[i])){
				return nil, errors.New("Invalid type of auxiliary data.")
			}
		}
	}

	// Use random k-1 degree polynomial to generate shares.
	shares := make([]*SecretShare, sss.participantCount)
	poly := sss.ShamirSecretSharingITF.GetRandomPolynomial(secret)
	for i:=0;i<sss.participantCount;i++{
		qr, err := poly.Calculate(auxiliary[i])
		if (err != nil) {return nil, err}
		value := NewShamirSecretShareValue(auxiliary[i],qr)
		shares[i] = NewSecretShare(i,value)
	}

	return shares, nil
}

/**
 * Calculating secret from input shares.
 *
 * @param shares The shares from which secret is calculated.
 * @return The secret calculated from the input shares.
 * @error If any of the input shares is invalid.
 */
func (sss *ShamirSecretSharing) calculateSecretImpl(shares []*SecretShare) (interface{},error){
    for i:=0;i<len(shares);i++{
    	valueGet := shares[i].GetValue()
    	valueReal,ok := valueGet.(*ShamirSecretShareValue)
    	if (!ok) {return nil, errors.New("Invalid type of share value, should be ShamirSecretShareValue.")}
    	if (!sss.ShamirSecretSharingITF.checkElement(valueReal.GetR()) || !sss.ShamirSecretSharingITF.checkElement(valueReal.GetQr())){
    		return nil, errors.New("Invalid type of elements in ShamirSecretShareValue.")
		}
	}
	// Since it matches threshold access structure, at least k equations are provided.
	// Use the first k equations to solve the solution.
	threshold := sss.access.GetThreshold()
	linearEquationSystem := sss.ShamirSecretSharingITF.GetEquationSystem()
	for i := 0; i < threshold ; i++{
		value := shares[i].GetValue().(*ShamirSecretShareValue)
		err := linearEquationSystem.
			AddEquation(sss.ShamirSecretSharingITF.GetEquationCoefficients(value.GetR()) , value.GetQr())
		if (err != nil) {return nil, err}
	}
	solution, err := linearEquationSystem.Solve()
	if (err != nil) {return nil, errors.New("At least one invalid share value.")}
	return solution[0],nil
}

/**
 * Set access structure.
 * <p>
 * Currently only <code>ThresholdAccessStructure</code> is accepted.
 *
 * @param access Access Structure.
 * @return error If the input access structure is invalid.
 */
func (sss *ShamirSecretSharing) SetAccessStructure(access AccessStructureInterface) error{
	accessValue, ok := access.(*ThresholdAccessStructure)
	if (!ok) {return errors.New("Invalid AccessStructure type. Should be 'ThresholdAccessStructure'.")}
	if (accessValue.participantCount != sss.participantCount) {
		return errors.New("The participant count in AccessStructure and Scheme should be equal.")
	}
	sss.access = accessValue
	return nil
}