describe('Unified Mode', () => {
    before(() => {
        cy.clear_db();
		cy.donate_image('apple1.jpeg');
    })

    it('Browse Image', () => {
		/*cy.visit('http://127.0.0.1:8080/donate');

        cy.fixture('images/apples/apple1.jpeg').then(fileContent => {
            cy.get('[id="dropzone"]').attachFile({
                fileContent: fileContent.toString(),
                fileName: 'apple1.jpeg',
                mimeType: 'image/png'
            });
        });*/
    });
})
