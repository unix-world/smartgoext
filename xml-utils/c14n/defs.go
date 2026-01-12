package c14n

// imported by unixman, only required things


type AlgorithmID string

func (id AlgorithmID) String() string {
	return string(id)
} //END FUNCTION


const ( // Supported canonicalization algorithms
	CanonicalXML10ExclusiveAlgorithmId 				AlgorithmID = "http://www.w3.org/2001/10/xml-exc-c14n#"
	CanonicalXML10ExclusiveWithCommentsAlgorithmId 	AlgorithmID = "http://www.w3.org/2001/10/xml-exc-c14n#WithComments"

	CanonicalXML11AlgorithmId 						AlgorithmID = "http://www.w3.org/2006/12/xml-c14n11"
	CanonicalXML11WithCommentsAlgorithmId 			AlgorithmID = "http://www.w3.org/2006/12/xml-c14n11#WithComments"

	CanonicalXML10RecAlgorithmId 					AlgorithmID = "http://www.w3.org/TR/2001/REC-xml-c14n-20010315"
	CanonicalXML10WithCommentsAlgorithmId 			AlgorithmID = "http://www.w3.org/TR/2001/REC-xml-c14n-20010315#WithComments"
)

// #end

