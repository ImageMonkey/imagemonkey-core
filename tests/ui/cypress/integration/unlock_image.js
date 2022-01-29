describe('Unlock Image', () => {
    before(() => {
        cy.clear_db_and_create_moderator_account();
		cy.donate_image('apple1.jpeg');
		cy.donate_image('apple2.jpeg');
	})

    it('Unlock Image', () => {	
		cy.login('moderator', 'moderator');
    	cy.unlock_all_images();
	});
})
