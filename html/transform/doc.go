/*
Package transform implements a html css selector and transformer.

An html doc can be inspected and queried using a subset of css selectors as
well as transformed.

	tree, _ := h5.New(rdr)
	t := transform.New(Tree)
	t.Apply(CopyAnd(myModifiers...), "li.menuitem")
	t.Apply(Replace(Text("my new text"), "a")

You can use the building blocks in this package as an html templating engine.
the basic principle behind html transform is that your template is just data
the same as your fill data. You use functions to mix and munge these two types
of data and get a third type of data back out which you can use to render.

You can also use this package to slice up and retrieve data out of an html page.
You can build scrapers and even mash up two different html documents.

How to do common templating actions.

How do I loop over input data?

Just because you don't have a for statement in your templating language doesn't
mean you can't loop over your data. You just have to loop over before you
insert it.

If we have template that looks like this:

   <html><body><ul id="people"><li>some guy</ul></body></html>

Then we can turn that list of people into a list of Me, Myself, and I like so:

   // We are going to want to set up the transformations for each person
   // in our list.
   liTransfroms := []TransformFuncs
   for i, item := range []string{"Me", "Myself", "I"} {
       liTransforms = append(menuTransforms, ReplaceChildren(Text(item)))
   }
   // Find the li in the ul element with the id='people' and make one copy of
   // it for each item from above with the text content replaced with that
   // item
    t.Apply(CopyAnd(liTransforms...), "ul#people", "li")

How do I change an html elements contents?

   t.Apply(ReplaceChildren(Text("some text contents"), "#SomeElement")

How do I remove/replace an html element.

  // Remove the element
  T.Apply(Replace(), "#SomeElement")
  // Replace the element
  node, _ := NewDoc("<span>hello</span>")
  T.Apply(Replace(node), "#SomeElement")
*/
package transform
