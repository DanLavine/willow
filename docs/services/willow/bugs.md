List of bugs that need to be addressed

1. Setup a blocking request trying to dequeue an item from a channel when there is nothing to remove.
   Then trigger a delete command to the queue. This should unblock the client waiting to dequeue with
   a proper error message back, but the Delete hangs untill the original dequeue clients are canceled