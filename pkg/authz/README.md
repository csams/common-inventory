# Authorization

I'd like to see something here similar to the abstractions in authn if possible so we can configure the
controllers.  We know we'll need something for enforcement that delegates to
[kessel](https://github.com/project-kessel), but those APIs and implementation are still evolving.  We need to
work with Alec Henninger and team to see what this would look like.

It currently has two authorizer implementations: one that allows full access and another that can delegate to
a service hosting the Kessel relations-api.
