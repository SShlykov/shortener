package app

func (app *App) appCheckers() []DependencyChecker {
	return []DependencyChecker{
		// Что-то вроде checkers.NewPostgresChecker(app.db),
		// Что-то вроде checkers.NewAPIChecker(link),
	}
}
