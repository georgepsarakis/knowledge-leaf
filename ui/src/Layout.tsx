import {
    Image,
    Container,
    Heading,
    Link,
    LinkBox,
    LinkOverlay,
    Flex,
    Separator
} from "@chakra-ui/react"

function Layout() {
    return <Container fluid marginTop={2} marginBottom={5}>
        <Flex justifyContent="space-between" gap="4">
            <LinkBox padding={1}>
                <LinkOverlay asChild>
                    <Image height={10} src="/logo192.png" />
                </LinkOverlay>
            </LinkBox>
            <Flex flexGrow={1} justify={"flex-end"}>
            <LinkBox padding={4}>
                <LinkOverlay asChild>
                    <Link href={"/"}>
                        <Heading color={"teal.500"} size={"md"}>Random</Heading>
                    </Link>
                </LinkOverlay>
            </LinkBox>
            <LinkBox padding={4}>
                <LinkOverlay asChild>
                    <Link href={"/on-this-day/events"}>
                        <Heading color={"teal.500"} size={"md"}>On This Day</Heading></Link>
                </LinkOverlay>
            </LinkBox>
            </Flex>
        </Flex>
        <Separator/>
    </Container>;
}

export default Layout;