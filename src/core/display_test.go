package core_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "helena/core"
)

// import { Script } from "./syntax";
// import { FALSE, INT, REAL, ScriptValue, STR, TRUE } from "./values";

var _ = Describe("Display", func() {
	Describe("UndisplayableValue", func() {
		It("should generate a placeholder block with a comment", func() {
			Expect(UndisplayableValue()).To(Equal("{#{undisplayable value}#}"))
		})
		It("should accept a custom comment", func() {
			Expect(UndisplayableValueWithLabel("some value")).To(Equal("{#{some value}#}"))
		})
	})
	//
	Describe("DisplayLiteralOrString", func() {
		It("should generate an empty quoted string for an empty string", func() {
			Expect(DisplayLiteralOrString("")).To(Equal(`""`))
		})
		It("should generate a literal for a string with no special character", func() {
			Expect(DisplayLiteralOrString("name")).To(Equal("name"))
			Expect(DisplayLiteralOrString("nameWithUnicode\u1234")).To(Equal("nameWithUnicode\u1234"))
		})
		It("should generate a quoted string for strings with special characters", func() {
			Expect(DisplayLiteralOrString("some value")).To(Equal(`"some value"`))
			Expect(DisplayLiteralOrString("$value")).To(Equal(`"\$value"`))
			Expect(DisplayLiteralOrString(`\#{[()]}#"`)).To(Equal(`"\\#\{\[\()]}#\""`))
		})
	})
	//
	Describe("DisplayLiteralOrBlock", func() {
		It("should generate an empty block for an empty string", func() {
			Expect(DisplayLiteralOrBlock("")).To(Equal("{}"))
		})
		It("should generate a literal for a string with no special character", func() {
			Expect(DisplayLiteralOrBlock("name")).To(Equal("name"))
			Expect(DisplayLiteralOrBlock("nameWithUnicode\u1234")).To(Equal("nameWithUnicode\u1234"))
		})
		It("should generate a quoted string for strings with special characters", func() {
			Expect(DisplayLiteralOrBlock("some value")).To(Equal("{some value}"))
			Expect(DisplayLiteralOrBlock("$value")).To(Equal("{\\$value}"))
			Expect(DisplayLiteralOrBlock(`{\#{[()]}#"}`)).To(Equal(`{\{\\\#\{\[\(\)\]\}\#\"\}}`))
		})
	})

	//	Describe("displayList", func () {
	//	  It("should generate a whitespace-separated list of values", func () {
	//	    const values = [STR("literal"), STR("some string"), INT(123), REAL(1.23)];
	//	    Expect(displayList(values)).To(Equal(`literal "some string" 123 1.23`);
	//	  });
	//	  It("should replace non-displayable values with placeholder", func () {
	//	    const values = [TRUE, new ScriptValue(new Script(), undefined), FALSE];
	//	    Expect(displayList(values)).To(Equal(
	//	      "true {#{undisplayable value}#} false"
	//	    );
	//	  });
	//	  It("should accept custom display function for non-displayable values", func () {
	//	    const values = [TRUE, new ScriptValue(new Script(), undefined), FALSE];
	//	    Expect(displayList(values, func () undisplayableValue("foo"))).To(Equal(
	//	      "true {#{foo}#} false"
	//	    );
	//	  });
	// })
})
