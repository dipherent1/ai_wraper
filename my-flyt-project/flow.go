package main

import (
	"github.com/mark3labs/flyt"
)

// CreateQAFlow creates a question-answering flow
func CreateQAFlow() *flyt.Flow {
	// Create nodes
	getQuestionNode := CreateGetQuestionNode()
	answerNode := CreateAnswerNode()

	// Connect nodes in sequence
	flow := flyt.NewFlow(getQuestionNode)
	flow.Connect(getQuestionNode, flyt.DefaultAction, answerNode)

	return flow
}

// CreateAgentFlow creates a more complex agent flow with decision making
func CreateAgentFlow() *flyt.Flow {
	// Create nodes
	analyzeNode := CreateAnalyzeNode()
	searchNode := CreateSearchNode()
	processNode := CreateProcessNode()
	answerNode := CreateAnswerNode()

	// Create flow with conditional routing
	flow := flyt.NewFlow(analyzeNode)

	// Connect based on analysis results
	flow.Connect(analyzeNode, "search", searchNode)
	flow.Connect(analyzeNode, "process", processNode)
	flow.Connect(analyzeNode, "answer", answerNode)

	// Search can lead back to analyze or to process
	flow.Connect(searchNode, "analyze", analyzeNode)
	flow.Connect(searchNode, "process", processNode)

	// Process always leads to answer
	flow.Connect(processNode, flyt.DefaultAction, answerNode)

	return flow
}

// CreateBatchFlow creates a flow that processes multiple items
func CreateBatchFlow() *flyt.Flow {
	// Create nodes
	loadItemsNode := CreateLoadItemsNode()
	batchProcessNode := CreateBatchProcessNode()
	aggregateNode := CreateAggregateResultsNode()

	// Connect nodes
	flow := flyt.NewFlow(loadItemsNode)
	flow.Connect(loadItemsNode, flyt.DefaultAction, batchProcessNode)
	flow.Connect(batchProcessNode, flyt.DefaultAction, aggregateNode)

	return flow
}
