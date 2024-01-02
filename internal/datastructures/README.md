bTrees
------

For all the bTrees (bTree, bTreeAssociated, bTreeOneToMany), there are some naming conventions on the functions
and their expected behaviors:

1. Delete
  * deleting from the tree is a blocking operation for any calls that want to interact with the same item
  * it is expected that if an item is deleted, then it is reasonable for the item to be recreated easily
1. Destroy
  * when destroying an item, any other request that wants to interact with the item is rejected and recieves an error
    that the item is being destroyed.
  * it is expected that this can be a long operation, but the caller would need to make an explicit(s) call to recreate
    each of the destroyed item(s)