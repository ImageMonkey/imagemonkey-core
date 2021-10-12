describe('Signup', () => {
    before(() => {
        cy.clear_db();
    })

    it('Testing user signup', () => {
    	cy.signup_user('user1', 'a', 'a', 'user1@imagemonkey.io');
	});
})
