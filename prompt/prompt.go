package prompt

import "github.com/AlecAivazis/survey/v2"

var AskSurvey = func(qs []*survey.Question, response interface{}, opts ...survey.AskOpt) error {
	return survey.Ask(qs, response, opts...)
}
