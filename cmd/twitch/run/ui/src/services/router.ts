import { Router } from '@vaadin/router';

const router = new Router();

const setRoute = (path: string) => {
	router.render(path);
	history.pushState({}, '', path);
};

export { router, setRoute };
