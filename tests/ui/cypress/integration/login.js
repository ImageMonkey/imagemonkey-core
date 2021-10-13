describe('Login', () => {
    before(() => {
        cy.clear_db();
        cy.signup_user('user', 'password', 'user@imagemonkey.io');
    })

    it('Testing user login', () => {
        cy.login('user', 'password');
    });
})