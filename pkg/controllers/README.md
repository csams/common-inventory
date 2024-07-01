# MVC Controllers

These are controllers in the MVC sense, not Kubernetes.

The controllers are placeholders while we figure out what the models are going to look like and just how
generic the resource type handling is going to be.  The code currently represents the extreme of a different
controller per type with explicit routes for the CRUD operations of each type.

I would like for these to be generic over a small number of kinds of model objects if possible: maybe one for
models that are resource types and one for those that are non-resource types.

They also might need to publish to kafka with a consumer pulling from kafka before inserting into the
database, but we'll see.  If authz APIs are defined as part of this API surface, we may require synchronous
APIs anyway.
