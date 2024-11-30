import {Center, Container, Heading, HStack, Link, LinkBox, LinkOverlay, Separator} from "@chakra-ui/react"

function Layout() {
    return <Container maxW={"2xl"} marginTop={2} marginBottom={5}>
        <Center>
            <HStack>
                <LinkBox padding={4}>
                    <LinkOverlay asChild>
                        <Link href={"/"}>
                            <Heading color={"teal.500"} size={"xl"}>Random</Heading>
                        </Link>
                    </LinkOverlay>
                </LinkBox>
                <LinkBox padding={4}>
                    <LinkOverlay asChild>
                        <Link href={"/on-this-day/events"}>
                            <Heading color={"teal.500"} size={"xl"}>On This Day</Heading></Link>
                    </LinkOverlay>
                </LinkBox>
            </HStack>
        </Center>
    </Container>;
}

export default Layout;