# One goal per category

Goal rows are unique on (budget_id, category_id), so a category holds at most one goal at a time; a finished goal must be deleted before a new one is created.
