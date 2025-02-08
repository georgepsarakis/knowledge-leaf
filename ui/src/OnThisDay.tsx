import {
    Center, ChakraProvider, Container, defaultSystem, Image,
    Heading, Spinner, Text, Box, Button, Link, Float, ListRoot, ListItem, Stack
} from "@chakra-ui/react";
import { Swiper, SwiperSlide } from 'swiper/react';
import 'swiper/css';
import "./index.css";
import {useParams} from "react-router";
import React, {useEffect, useState} from "react";
import {getDomain} from "./backend/backend";
import {Navigation} from 'swiper/modules';
import 'swiper/css/navigation';
import {AiOutlineArrowLeft, AiOutlineArrowRight} from "react-icons/ai";
import Layout from "./Layout";
import { LuExternalLink } from "react-icons/lu";
import { ClipboardIconButton, ClipboardRoot } from "./components/ui/clipboard";



type ArticleImage = {
    url: string;
    height: number;
    width: number;
}

type OnThisDayEvent = {
    description: string;
    title: string;
    year: number;
    short_title: string;
    extract: string;
    image: ArticleImage;
    url: string;
    references: OnThisDayEventReference[];
    app_link_url: string;
}

type OnThisDayEventReference = {
    url: string;
    title: string;
}

type OnThisDayEventsResponse = {
    titles: OnThisDayEvent[];
}

function AppLinkURL(props: {url: string}) {
    return  <ClipboardRoot value={props.url}>
                <ClipboardIconButton />
            </ClipboardRoot>
}

function buildAPIRequestURL(date?: string | null, title?: string | null): string {
    date ??= null;
    title ??= null;
    const baseURL = getDomain()+'/on-this-day/events';
    if (date === null || title === null) {
        return baseURL;
    }
    return `${baseURL}/${encodeURIComponent(date)}/${encodeURIComponent(title)}`;
}

function EventsOnThisDay() {
    const [data, setData] = useState<OnThisDayEventsResponse>();
    const [isLoading, setIsLoading] = useState<boolean>(true)
    let ignore = false;
    let { date, title } = useParams();
    useEffect(() => {
        if (!ignore) {
            ignore = true;
            fetch(buildAPIRequestURL(date, title)).then(response => {
                return response.json()
            }).then(triviaData => {
                setData(triviaData);
                setIsLoading(false);
            }).catch((error) => {
                console.log(error);
            });
        }
        return () => {
            ignore = true;
        }
    }, []);
    return <ChakraProvider value={defaultSystem}>
        <Layout/>
        <Center>
                <Container marginTop={5} maxW={"4xl"}>
                    <Box>
                        <Float placement={"top-start"}>
                            <Button marginTop={9} backgroundColor="teal" className={"kl-swiper-button-prev"} size={"xl"}>
                                <AiOutlineArrowLeft/>
                            </Button>
                        </Float>
                    </Box>
                    <Center>
                        <Heading size={"2xl"} color={"gray.500"}>On This Day</Heading>
                    </Center>
                            {
                                isLoading ?
                                    <Center>
                                        <Spinner size={"xl"} color={"teal.500"} marginTop={20}/>
                                    </Center> :
                                        <Swiper
                                            modules={[Navigation]}
                                            navigation={{
                                                nextEl: '.kl-swiper-button-next',
                                                prevEl: '.kl-swiper-button-prev',
                                            }}
                                            slidesPerView={1}>
                                            {
                                                data!.titles.map((t) => {
                                                    const key = t.url + "-" + t.year.toString();
                                                    const dt = new Date();
                                                    dt.setFullYear(t.year);
                                                    const displayDate = dt.toLocaleDateString(
                                                        undefined,
                                                        {month: 'long', year: "numeric", day: "numeric"});
                                                    const displayTitle = `${displayDate}: ${t.title}`;
                                                    return <SwiperSlide key={key}>
                                                        <Box p="4"
                                                             borderWidth="1px"
                                                             borderColor="border.disabled"
                                                             color="fg.disabled" margin={30}>
                                                        <Event
                                                            key={key}
                                                            title={displayTitle}
                                                            extract={t.extract}
                                                            image={t.image}
                                                            short_title={t.short_title}
                                                            description={t.description}
                                                            url={t.url}
                                                            references={t.references}
                                                            app_link_url={window.location.host + t.app_link_url}
                                                        />
                                                        </Box>
                                                    </SwiperSlide>
                                                })
                                            }
                                        </Swiper>
                            }
                    <Float placement={"top-end"}>
                        <Button marginTop={9} backgroundColor="teal" className={"kl-swiper-button-next"} size={"xl"}>
                            <AiOutlineArrowRight/>
                        </Button>
                    </Float>
                </Container>

        </Center>
       </ChakraProvider>;
}

type EventProps = {
    title: string;
    extract: string;
    image: ArticleImage;
    short_title: string;
    description: string;
    url: string;
    references: OnThisDayEventReference[];
    app_link_url: string;
}

function Event(props: EventProps) {
    const image = props.image;
    return <Container maxW={"2xl"}>

        <Stack key={"event-heading-"+props.title} marginBottom={5}>
             <Heading position="relative" size={"lg"}>
                 {props.title}
                 <Float placement={"middle-start"}>
                     <Box paddingRight={50}>
                        <AppLinkURL url={props.app_link_url}/>
                     </Box>
                 </Float>
             </Heading>
        </Stack>
             <Center>
                 {image.url ?
                    <Image src={image.url}
                           width={image.width}
                           height={image.height}
                           transition={"0.5s linear"} marginBottom={5}/> : ''}
             </Center>
             <Text marginBottom={5}>{props.extract}</Text>
             <Link target={"_blank"} href={props.url}>
                <Text color={"teal"}>{props.short_title} - {props.description}</Text>
                 <LuExternalLink color={"teal"}/>
             </Link>
             <RenderReferences items={props.references}/>
           </Container>;
}

type ReferencesProps = {
    items: OnThisDayEventReference[];
}
function RenderReferences({ items }: ReferencesProps) {
    if (items === null || items.length === 0) {
        return null;
    }

    return <Box marginTop={2}>
        <Heading size={"md"} marginBottom={2}>References</Heading>
        <ListRoot>
            {items.map((ref) =>
                <ListItem key={ref.url}>
                    <Link target={"_blank"} href={ref.url}>
                        <Text color={"teal"}>{ref.title}</Text><LuExternalLink color={"teal"}/>
                    </Link>
                </ListItem>
            )}
        </ListRoot>
    </Box>
}

export default EventsOnThisDay;